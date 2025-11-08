package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/emersion/go-ical"
)

// writeCalendarToFile handles the common pattern of creating a file and encoding a calendar
func writeCalendarToFile(calendar *ical.Calendar, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	encoder := ical.NewEncoder(file)
	return encoder.Encode(calendar)
}

// readCalendarFromFile handles opening and decoding a calendar file
func readCalendarFromFile(filePath string) (*ical.Calendar, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	decoder := ical.NewDecoder(file)
	calendar, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to parse .ics file %s: %w", filePath, err)
	}

	return calendar, nil
}

// getBaseDir returns the base directory for calendar storage
func getBaseDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "chronos", "calendars")
}

// getCalendarDir returns the directory path for a calendar
func getCalendarDir(calendarID string) string {
	return filepath.Join(getBaseDir(), calendarID)
}

// getEventPath returns the full path for an individual event file
func getEventPath(calendarID string, eventUID string) string {
	return filepath.Join(getCalendarDir(calendarID), fmt.Sprintf("%s.ics", eventUID))
}

// ensureBaseDir creates the base directory structure if it doesn't exist
func ensureBaseDir() error {
	return os.MkdirAll(getBaseDir(), 0o755)
}

// ensureCalendarDir creates the calendar directory if it doesn't exist
func ensureCalendarDir(calendarID string) error {
	return os.MkdirAll(getCalendarDir(calendarID), 0o755)
}

// createSingleEventCalendar wraps a single event in a VCALENDAR container
// Required by iCalendar RFC 5545 - VEVENT must be inside VCALENDAR
func createSingleEventCalendar(event *ical.Event) *ical.Calendar {
	calendar := ical.NewCalendar()
	calendar.Props.SetText(ical.PropVersion, "2.0")
	calendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")
	calendar.Children = []*ical.Component{event.Component}
	return calendar
}