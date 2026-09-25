package storage

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/samuelstranges/chronos/util"
)

// CalDAVClient wraps the go-webdav CalDAV client with convenience methods
type CalDAVClient struct {
	client       *caldav.Client
	rawClient    webdav.HTTPClient
	serverURL    *url.URL
	config       *CalDAVConfig
	homeSet      string                      // Calendar home set path
	calendars    map[string]string           // calendarID → server path
	calendarInfo map[string]*caldav.Calendar // calendarID → calendar metadata
}

// NewCalDAVClient creates and initializes a CalDAV client
func NewCalDAVClient(config *CalDAVConfig) (*CalDAVClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create HTTP client with timeout and Basic Auth
	baseClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpClient := webdav.HTTPClientWithBasicAuth(
		baseClient,
		config.Username,
		config.Password,
	)

	// Create CalDAV client
	client, err := caldav.NewClient(httpClient, config.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create CalDAV client: %w", err)
	}

	parsedServerURL, err := url.Parse(config.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	cc := &CalDAVClient{
		client:       client,
		rawClient:    httpClient,
		serverURL:    parsedServerURL,
		config:       config,
		calendars:    make(map[string]string),
		calendarInfo: make(map[string]*caldav.Calendar),
	}

	// Discover calendar home set
	if err := cc.discoverCalendars(); err != nil {
		return nil, fmt.Errorf("failed to discover calendars: %w", err)
	}

	return cc, nil
}

// discoverCalendars finds all calendars under the configured calendar home
func (cc *CalDAVClient) discoverCalendars() error {
	ctx := context.Background()

	// Use the configured calendar home URL directly (no auto-discovery)
	calendarHomeURL := cc.config.CalendarHomeURL
	if calendarHomeURL == "" {
		return fmt.Errorf("calendar_home_url is required in config")
	}

	cc.homeSet = calendarHomeURL

	// Find all calendars under this home
	calendars, err := cc.client.FindCalendars(ctx, calendarHomeURL)
	if err != nil {
		return fmt.Errorf("failed to find calendars at %s: %w", calendarHomeURL, err)
	}

	// Map calendar IDs to server paths
	for _, cal := range calendars {
		// Use calendar name as ID (sanitized for filesystem safety)
		calendarID := sanitizeCalendarID(cal.Name)
		cc.calendars[calendarID] = cal.Path
		cc.calendarInfo[calendarID] = &cal
	}

	return nil
}

// GetCalendarIDs returns a list of all calendar IDs
func (cc *CalDAVClient) GetCalendarIDs() []string {
	ids := make([]string, 0, len(cc.calendars))
	for id := range cc.calendars {
		ids = append(ids, id)
	}
	return ids
}

// FetchCalendar retrieves all events from a calendar
//
// This issues the calendar-query REPORT itself (rather than using
// caldav.Client.QueryCalendar) because some servers - iCloud in particular -
// include response entries in the multistatus whose calendar-data property
// comes back with its own 404 status, distinct from the response's overall
// status. The upstream decoder treats any such per-property error as fatal
// for the whole batch; we instead skip just that entry and keep going.
func (cc *CalDAVClient) FetchCalendar(calendarID string) (*ical.Calendar, error) {
	calendarPath, exists := cc.calendars[calendarID]
	if !exists {
		return nil, fmt.Errorf("calendar '%s' not found", calendarID)
	}

	ctx := context.Background()

	objects, skipped, err := cc.queryCalendarObjects(ctx, calendarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to query calendar: %w", err)
	}
	for _, s := range skipped {
		util.LogErrorToFile(fmt.Sprintf("caldav: skipping unreadable event %q in calendar %q: %s", s.href, calendarID, s.reason))
	}

	// Create a merged calendar from all objects
	calendar := ical.NewCalendar()
	calendar.Props.SetText(ical.PropVersion, "2.0")
	calendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")

	// Add calendar name if available
	if info, exists := cc.calendarInfo[calendarID]; exists {
		calendar.Props.SetText("X-WR-CALNAME", info.Name)
		if info.Description != "" {
			calendar.Props.SetText("X-WR-CALDESC", info.Description)
		}
	}

	// Merge all events into one calendar
	for _, obj := range objects {
		for _, child := range obj.Children {
			if child.Name == "VEVENT" {
				calendar.Children = append(calendar.Children, child)
			}
		}
	}

	return calendar, nil
}

// skippedObject records a multistatus response entry that couldn't be read as an event
type skippedObject struct {
	href   string
	reason string
}

// caldavMultistatus is a minimal, lenient decode target for a calendar-query REPORT response
type caldavMultistatus struct {
	XMLName   xml.Name         `xml:"DAV: multistatus"`
	Responses []caldavResponse `xml:"DAV: response"`
}

type caldavResponse struct {
	Href      string           `xml:"DAV: href"`
	Propstats []caldavPropstat `xml:"DAV: propstat"`
}

type caldavPropstat struct {
	Status string     `xml:"DAV: status"`
	Prop   caldavProp `xml:"DAV: prop"`
}

type caldavProp struct {
	CalendarData string `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
}

// calendarData returns the event body from the propstat block whose status is 2xx, if any
func (r caldavResponse) calendarData() (string, bool) {
	for _, ps := range r.Propstats {
		if strings.Contains(ps.Status, " 2") && ps.Prop.CalendarData != "" {
			return ps.Prop.CalendarData, true
		}
	}
	return "", false
}

// queryCalendarObjects performs a calendar-query REPORT for all VEVENTs under calendarPath,
// returning the successfully-parsed calendar objects and a list of entries that had to be skipped.
func (cc *CalDAVClient) queryCalendarObjects(ctx context.Context, calendarPath string) ([]*ical.Calendar, []skippedObject, error) {
	const reqBody = `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>
    <D:getetag/>
    <C:calendar-data/>
  </D:prop>
  <C:filter>
    <C:comp-filter name="VCALENDAR">
      <C:comp-filter name="VEVENT"/>
    </C:comp-filter>
  </C:filter>
</C:calendar-query>`

	httpReq, err := http.NewRequestWithContext(ctx, "REPORT", cc.resolveURL(calendarPath), strings.NewReader(reqBody))
	if err != nil {
		return nil, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/xml; charset=utf-8")
	httpReq.Header.Set("Depth", "1")

	resp, err := cc.rawClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, nil, fmt.Errorf("unexpected status %s: %s", resp.Status, string(body))
	}

	var ms caldavMultistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, nil, fmt.Errorf("failed to decode multistatus response: %w", err)
	}

	var calendars []*ical.Calendar
	var skipped []skippedObject
	for _, r := range ms.Responses {
		data, ok := r.calendarData()
		if !ok {
			skipped = append(skipped, skippedObject{href: r.Href, reason: "calendar-data property unavailable"})
			continue
		}

		cal, err := ical.NewDecoder(strings.NewReader(data)).Decode()
		if err != nil {
			skipped = append(skipped, skippedObject{href: r.Href, reason: fmt.Sprintf("failed to parse iCal data: %v", err)})
			continue
		}

		calendars = append(calendars, cal)
	}

	return calendars, skipped, nil
}

// resolveURL builds an absolute request URL for an absolute server path
func (cc *CalDAVClient) resolveURL(p string) string {
	u := url.URL{
		Scheme: cc.serverURL.Scheme,
		User:   cc.serverURL.User,
		Host:   cc.serverURL.Host,
		Path:   p,
	}
	return u.String()
}

// SaveEvent uploads a single event to the CalDAV server
func (cc *CalDAVClient) SaveEvent(calendarID string, event *ical.Event) (string, error) {
	calendarPath, exists := cc.calendars[calendarID]
	if !exists {
		return "", fmt.Errorf("calendar '%s' not found", calendarID)
	}

	// Get event UID
	eventUID := event.Props.Get(ical.PropUID)
	if eventUID == nil {
		return "", fmt.Errorf("event has no UID")
	}

	// Construct event path on server
	eventPath := fmt.Sprintf("%s%s.ics", calendarPath, eventUID.Value)

	// Create a single-event calendar for upload (required by CalDAV protocol)
	singleEventCal := createSingleEventCalendar(event)

	ctx := context.Background()

	// Upload to server
	obj, err := cc.client.PutCalendarObject(ctx, eventPath, singleEventCal)
	if err != nil {
		return "", fmt.Errorf("failed to upload event to path %s: %w", eventPath, err)
	}

	// Return ETag for change tracking
	return obj.ETag, nil
}

// DeleteEvent removes a single event from the CalDAV server
func (cc *CalDAVClient) DeleteEvent(calendarID string, eventUID string) error {
	calendarPath, exists := cc.calendars[calendarID]
	if !exists {
		return fmt.Errorf("calendar '%s' not found", calendarID)
	}

	// Construct event path on server
	eventPath := fmt.Sprintf("%s%s.ics", calendarPath, eventUID)

	ctx := context.Background()

	// Delete from server using the CalDAV client
	err := cc.client.RemoveAll(ctx, eventPath)
	if err != nil {
		return fmt.Errorf("failed to delete event from path %s: %w", eventPath, err)
	}

	return nil
}

// sanitizeCalendarID converts a calendar name to a filesystem-safe ID
func sanitizeCalendarID(name string) string {
	// Replace spaces with underscores, convert to lowercase
	id := strings.ReplaceAll(name, " ", "_")
	id = strings.ToLower(id)

	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()

	// Ensure we don't return empty string
	if sanitized == "" {
		sanitized = "calendar"
	}

	return sanitized
}
