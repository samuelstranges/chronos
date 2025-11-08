package storage

import (
	"fmt"

	"github.com/emersion/go-ical"
)

// CalDAVStorage implements CalendarStorage using a remote CalDAV server
// All operations go directly to the server - no local caching
type CalDAVStorage struct {
	client *CalDAVClient
	config *CalDAVConfig
}

// NewCalDAVStorage creates a new CalDAV-based storage backend
func NewCalDAVStorage(config *CalDAVConfig) (*CalDAVStorage, error) {
	// Create CalDAV client
	client, err := NewCalDAVClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create CalDAV client: %w", err)
	}

	return &CalDAVStorage{
		client: client,
		config: config,
	}, nil
}

// LoadCalendars loads all calendars from the CalDAV server
func (cs *CalDAVStorage) LoadCalendars() (map[string]*ical.Calendar, error) {
	calendars := make(map[string]*ical.Calendar)

	// Get all calendar IDs from server
	calendarIDs := cs.client.GetCalendarIDs()

	// Fetch each calendar
	for _, calendarID := range calendarIDs {
		calendar, err := cs.client.FetchCalendar(calendarID)
		if err != nil {
			// Skip calendars that fail to load, but log the error
			continue
		}

		calendars[calendarID] = calendar
	}

	return calendars, nil
}

// SaveCalendars not supported for CalDAV
func (cs *CalDAVStorage) SaveCalendars(calendarMap map[string]*ical.Calendar) error {
	return fmt.Errorf("SaveCalendars not supported for CalDAV")
}

// SaveEvent saves a single event to the CalDAV server
func (cs *CalDAVStorage) SaveEvent(calendarID string, event *ical.Event) (string, error) {
	etag, err := cs.client.SaveEvent(calendarID, event)
	if err != nil {
		return "", fmt.Errorf("failed to save event: %w", err)
	}
	return etag, nil
}

// DeleteEvent deletes a single event from the CalDAV server
func (cs *CalDAVStorage) DeleteEvent(calendarID string, eventUID string) error {
	return cs.client.DeleteEvent(calendarID, eventUID)
}

// ExportCalendar exports a calendar to a local file
func (cs *CalDAVStorage) ExportCalendar(calendar *ical.Calendar, filePath string) error {
	// Ensure calendar is valid
	if err := ensureCalendarValid(calendar); err != nil {
		return err
	}

	return writeCalendarToFile(calendar, filePath)
}

// ExportAllCalendars exports all calendars to a merged file
func (cs *CalDAVStorage) ExportAllCalendars(calendars []*ical.Calendar, filePath string) error {
	// Create merged calendar
	mergedCalendar := ical.NewCalendar()
	mergedCalendar.Props.SetText(ical.PropVersion, "2.0")
	mergedCalendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")
	mergedCalendar.Props.SetText("X-WR-CALNAME", "Chronos Merged Export")

	// Merge all events
	for _, calendar := range calendars {
		for _, child := range calendar.Children {
			if child.Name == "VEVENT" {
				mergedCalendar.Children = append(mergedCalendar.Children, child)
			}
		}
	}

	// Validate and write
	if err := ensureCalendarValid(mergedCalendar); err != nil {
		return err
	}

	return writeCalendarToFile(mergedCalendar, filePath)
}

// ImportFromFile imports events from a local file
func (cs *CalDAVStorage) ImportFromFile(filePath string) (*ical.Calendar, error) {
	return readCalendarFromFile(filePath)
}

// SupportsSync returns true - CalDAV supports event-level operations
func (cs *CalDAVStorage) SupportsSync() bool {
	return true
}

// GetStorageType returns "caldav" to identify this storage backend
func (cs *CalDAVStorage) GetStorageType() string {
	return "caldav"
}
