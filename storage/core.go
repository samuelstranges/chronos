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

// SaveCalendars saves all calendars to disk as individual .ics files
func (fs *FileStorage) SaveCalendars(calendarMap map[string]*ical.Calendar) error {
	// Ensure base directory exists
	if err := ensureBaseDir(); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Export each calendar
	for calendarID, calendar := range calendarMap {
		// Ensure calendar is valid before writing
		if err := ensureCalendarValid(calendar); err != nil {
			return err
		}

		filePath := getCalendarPath(calendarID)
		if err := writeCalendarToFile(calendar, filePath); err != nil {
			return fmt.Errorf("failed to save calendar %s: %w", calendarID, err)
		}
	}

	return nil
}

// LoadCalendars loads all calendars from persistent storage
func (fs *FileStorage) LoadCalendars() (map[string]*ical.Calendar, error) {
	baseDir := getBaseDir()

	// Ensure base directory exists
	if err := ensureBaseDir(); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Read all .ics files in the directory (standard format)
	files, err := filepath.Glob(filepath.Join(baseDir, "*.ics"))
	if err != nil {
		return nil, fmt.Errorf("failed to list .ics files: %w", err)
	}

	calendars := make(map[string]*ical.Calendar)
	var skippedFiles []string

	for _, filePath := range files {
		// Extract calendar ID from filename (STABLE ID based on filename, not array index!)
		filename := filepath.Base(filePath)
		calendarID := filename[:len(filename)-4] // Remove .ics extension

		// Load the calendar using helper
		calendar, err := readCalendarFromFile(filePath)
		if err != nil {
			skippedFiles = append(skippedFiles, filePath)
			continue // Skip files that can't be loaded
		}

		calendars[calendarID] = calendar
	}

	// Report if files were skipped but don't fail the entire operation
	if len(skippedFiles) > 0 {
		return calendars, fmt.Errorf("loaded %d calendars, skipped %d files with errors: %v",
			len(calendars), len(skippedFiles), skippedFiles)
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
