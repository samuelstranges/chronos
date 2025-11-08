package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/util"
)

// FileStorage implements CalendarStorage using local file system
type FileStorage struct{}

// NewFileStorage creates a new file-based calendar storage
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// ExportCalendar exports a specific calendar to a file implementing CalendarStorage interface
func (fs *FileStorage) ExportCalendar(calendar *ical.Calendar, filePath string) error {
	// Ensure calendar is valid before exporting
	if err := ensureCalendarValid(calendar); err != nil {
		return err
	}

	return writeCalendarToFile(calendar, filePath)
}

// ExportAllCalendars exports all calendars to a single merged file implementing CalendarStorage interface
func (fs *FileStorage) ExportAllCalendars(calendars []*ical.Calendar, filePath string) error {
	// Create a new merged calendar
	mergedCalendar := ical.NewCalendar()
	mergedCalendar.Props.SetText(ical.PropVersion, "2.0")
	mergedCalendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")
	mergedCalendar.Props.SetText("X-WR-CALNAME", "Chronos Merged Export")

	// Add all events from all calendars
	for _, calendar := range calendars {
		util.ForEachEventInCalendar(calendar, func(event *ical.Component) bool {
			// Create a deep copy of the event to avoid modifying original
			eventCopy := util.DeepCopyComponent(event)
			mergedCalendar.Children = append(mergedCalendar.Children, eventCopy)
			return true // Continue iteration
		})
	}

	// Ensure merged calendar is valid before exporting
	if err := ensureCalendarValid(mergedCalendar); err != nil {
		return err
	}

	return writeCalendarToFile(mergedCalendar, filePath)
}

// ImportFromFile imports events from a .ics file and returns the calendar implementing CalendarStorage interface
func (fs *FileStorage) ImportFromFile(filePath string) (*ical.Calendar, error) {
	return readCalendarFromFile(filePath)
}

// LoadCalendars loads all calendars from persistent storage (per-event file structure)
func (fs *FileStorage) LoadCalendars() (map[string]*ical.Calendar, error) {
	baseDir := getBaseDir()

	// Ensure base directory exists
	if err := ensureBaseDir(); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Read all calendar directories
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar directories: %w", err)
	}

	calendars := make(map[string]*ical.Calendar)
	var skippedEvents []string

	for _, entry := range entries {
		// Skip non-directories (ignore any stray files)
		if !entry.IsDir() {
			continue
		}

		calendarID := entry.Name()

		// Create merged calendar for this calendar ID
		mergedCalendar := ical.NewCalendar()
		mergedCalendar.Props.SetText(ical.PropVersion, "2.0")
		mergedCalendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")

		// Load all event files from this calendar's directory
		calendarDir := getCalendarDir(calendarID)
		eventFiles, err := filepath.Glob(filepath.Join(calendarDir, "*.ics"))
		if err != nil {
			continue // Skip this calendar if we can't list files
		}

		// Read each event file and merge into calendar
		for _, eventFilePath := range eventFiles {
			// Read single-event calendar
			singleEventCal, err := readCalendarFromFile(eventFilePath)
			if err != nil {
				skippedEvents = append(skippedEvents, eventFilePath)
				continue
			}

			// Extract VEVENT components and add to merged calendar
			for _, child := range singleEventCal.Children {
				if child.Name == "VEVENT" {
					mergedCalendar.Children = append(mergedCalendar.Children, child)
				}
			}
		}

		calendars[calendarID] = mergedCalendar
	}

	// Report if events were skipped but don't fail the entire operation
	if len(skippedEvents) > 0 {
		return calendars, fmt.Errorf("loaded %d calendars, skipped %d event files with errors: %v",
			len(calendars), len(skippedEvents), skippedEvents)
	}

	return calendars, nil
}

// SaveEvent saves a single event to a file
func (fs *FileStorage) SaveEvent(calendarID string, event *ical.Event) (string, error) {
	// Ensure calendar directory exists
	if err := ensureCalendarDir(calendarID); err != nil {
		return "", fmt.Errorf("failed to create calendar directory: %w", err)
	}

	// Ensure event has UID
	if err := ensureEventUID(event); err != nil {
		return "", fmt.Errorf("event missing UID: %w", err)
	}

	eventUID := event.Props.Get(ical.PropUID).Value

	// Create single-event calendar wrapper
	singleEventCal := createSingleEventCalendar(event)

	// Validate before writing
	if err := ensureCalendarValid(singleEventCal); err != nil {
		return "", err
	}

	// Write to individual event file
	eventPath := getEventPath(calendarID, eventUID)
	if err := writeCalendarToFile(singleEventCal, eventPath); err != nil {
		return "", fmt.Errorf("failed to save event: %w", err)
	}

	return "", nil // No ETag for file storage
}

// DeleteEvent removes a single event file
func (fs *FileStorage) DeleteEvent(calendarID string, eventUID string) error {
	eventPath := getEventPath(calendarID, eventUID)

	// Delete the event file
	if err := os.Remove(eventPath); err != nil {
		// Ignore error if file doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete event: %w", err)
		}
	}

	return nil
}

// SupportsSync returns true - file storage now supports event-level operations
func (fs *FileStorage) SupportsSync() bool {
	return true
}

// GetStorageType returns "file" to identify this storage backend
func (fs *FileStorage) GetStorageType() string {
	return "file"
}
