package timezone

import (
	"fmt"
	"time"

	"github.com/emersion/go-ical"
)

// FormTimeDisplay converts an iCal datetime value to user-friendly local time strings for forms
// Returns separate date and time strings for form fields
func FormTimeDisplay(icalValue string) (dateStr, timeStr string, err error) {
	if icalValue == "" {
		return "", "", fmt.Errorf("empty datetime value")
	}

	localTime, err := ParseICalDateTime(icalValue)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse datetime for form display: %w", err)
	}

	dateStr = localTime.Format("2006-01-02")
	timeStr = localTime.Format("15:04")
	return dateStr, timeStr, nil
}

// FormDateDisplay converts an iCal date value to user-friendly date string for forms
func FormDateDisplay(icalValue string) (string, error) {
	if icalValue == "" {
		return "", fmt.Errorf("empty date value")
	}

	localTime, err := ParseICalDate(icalValue)
	if err != nil {
		return "", fmt.Errorf("failed to parse date for form display: %w", err)
	}

	return localTime.Format("2006-01-02"), nil
}

// ParseFormInput converts user form input (date + time strings) to LocalTime
// This is the safe way to handle user input, always interpreting as local time
func ParseFormInput(dateStr, timeStr string) (LocalTime, error) {
	if dateStr == "" {
		return LocalTime{}, fmt.Errorf("date cannot be empty")
	}

	if timeStr == "" {
		// Date only (all-day event)
		parsedTime, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return LocalTime{}, fmt.Errorf("invalid date format '%s': %w", dateStr, err)
		}
		return LocalTime{parsedTime}, nil
	}

	// Date and time
	return NewLocalTimeFromDateTime(dateStr, timeStr)
}

// ConvertFormInputToICalFormat converts form input to iCal format while preserving timezone context
// Uses the original iCal value to determine the correct output format
func ConvertFormInputToICalFormat(dateStr, timeStr, originalICalValue string) (string, error) {
	localTime, err := ParseFormInput(dateStr, timeStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse form input: %w", err)
	}

	// Use the timezone preservation logic
	return PreserveTimezoneFormat(localTime, originalICalValue), nil
}

// PopulateFormFromEvent extracts display-friendly values from an ical.Event for form population
// Returns separate date and time strings that can be shown to the user
func PopulateFormFromEvent(event *ical.Event) (startDate, startTime, endDate, endTime string, err error) {
	if event == nil {
		return "", "", "", "", fmt.Errorf("event is nil")
	}

	// Get start time
	dtStartProp := event.Props.Get("DTSTART")
	if dtStartProp == nil {
		return "", "", "", "", fmt.Errorf("event has no DTSTART property")
	}

	if valueParam := dtStartProp.Params.Get("VALUE"); valueParam == "DATE" {
		// All-day event - only return date
		startDate, err = FormDateDisplay(dtStartProp.Value)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to parse start date: %w", err)
		}

		// Check for end date
		dtEndProp := event.Props.Get("DTEND")
		if dtEndProp != nil {
			endDate, err = FormDateDisplay(dtEndProp.Value)
			if err != nil {
				return "", "", "", "", fmt.Errorf("failed to parse end date: %w", err)
			}
		}

		return startDate, "", endDate, "", nil
	}

	// Timed event
	startDate, startTime, err = FormTimeDisplay(dtStartProp.Value)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse start datetime: %w", err)
	}

	// Get end time
	dtEndProp := event.Props.Get("DTEND")
	if dtEndProp != nil {
		endDate, endTime, err = FormTimeDisplay(dtEndProp.Value)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to parse end datetime: %w", err)
		}
	}

	return startDate, startTime, endDate, endTime, nil
}

// SetEventDateTimeProperty safely sets a datetime property on an event while preserving timezone format
func SetEventDateTimeProperty(event *ical.Event, propName string, localTime LocalTime, originalValue string) {
	if event == nil {
		return
	}

	// Use SetDateTime which properly handles datetime values without VALUE=TEXT
	event.Props.SetDateTime(propName, localTime.Time)
}