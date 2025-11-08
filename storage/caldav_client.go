package storage

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

// CalDAVClient wraps the go-webdav CalDAV client with convenience methods
type CalDAVClient struct {
	client       *caldav.Client
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

	cc := &CalDAVClient{
		client:       client,
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
func (cc *CalDAVClient) FetchCalendar(calendarID string) (*ical.Calendar, error) {
	calendarPath, exists := cc.calendars[calendarID]
	if !exists {
		return nil, fmt.Errorf("calendar '%s' not found", calendarID)
	}

	ctx := context.Background()

	// Query all events in the calendar
	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name: "VCALENDAR",
			Comps: []caldav.CalendarCompRequest{
				{Name: "VEVENT"},
			},
		},
	}

	objects, err := cc.client.QueryCalendar(ctx, calendarPath, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query calendar: %w", err)
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
		for _, child := range obj.Data.Children {
			if child.Name == "VEVENT" {
				calendar.Children = append(calendar.Children, child)
			}
		}
	}

	return calendar, nil
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
