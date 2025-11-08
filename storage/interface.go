package storage

import (
	"github.com/emersion/go-ical"
)

// CalendarStorage defines the interface for calendar persistence operations
type CalendarStorage interface {
	// === BULK OPERATIONS (for initial load/export) ===

	// LoadCalendars loads all calendars from persistent storage
	LoadCalendars() (map[string]*ical.Calendar, error)

	// === EXPORT/IMPORT ===

	// ExportCalendar exports a specific calendar to a file path
	ExportCalendar(calendar *ical.Calendar, filePath string) error

	// ExportAllCalendars exports all calendars to a single merged file
	ExportAllCalendars(calendars []*ical.Calendar, filePath string) error

	// ImportFromFile imports events from a .ics file and returns the calendar
	ImportFromFile(filePath string) (*ical.Calendar, error)

	// === EVENT-LEVEL OPERATIONS (for incremental sync) ===

	// SaveEvent creates or updates a single event in a calendar
	// Returns the server's ETag (empty string for file storage)
	SaveEvent(calendarID string, event *ical.Event) (etag string, err error)

	// DeleteEvent removes a single event from a calendar by UID
	DeleteEvent(calendarID string, eventUID string) error

	// === SYNC CAPABILITIES ===

	// SupportsSync indicates if this storage backend supports incremental sync
	// File storage returns false, CalDAV storage returns true
	SupportsSync() bool

	// GetStorageType returns a string identifier for the storage type
	// e.g., "file", "caldav", "hybrid"
	GetStorageType() string
}