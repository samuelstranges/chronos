package timezone

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-ical"
)

// ParseICalDateTime safely parses an iCal datetime value into LocalTime
// Handles both UTC (20060102T150405Z) and local (20060102T150405) formats
func ParseICalDateTime(icalValue string) (LocalTime, error) {
	if icalValue == "" {
		return LocalTime{}, fmt.Errorf("empty datetime value")
	}

	var parsedTime time.Time
	var err error

	if strings.HasSuffix(icalValue, "Z") {
		// UTC format
		parsedTime, err = time.Parse("20060102T150405Z", icalValue)
		if err != nil {
			return LocalTime{}, fmt.Errorf("failed to parse UTC datetime '%s': %w", icalValue, err)
		}
		// Convert UTC to local for display
		return NewLocalTime(parsedTime), nil
	}
	
	// Local format (assume local timezone)
	parsedTime, err = time.ParseInLocation("20060102T150405", icalValue, time.Local)
	if err != nil {
		return LocalTime{}, fmt.Errorf("failed to parse local datetime '%s': %w", icalValue, err)
	}
	return LocalTime{parsedTime}, nil
}

// ParseICalDate safely parses an iCal date value (for all-day events)
func ParseICalDate(icalValue string) (LocalTime, error) {
	if icalValue == "" {
		return LocalTime{}, fmt.Errorf("empty date value")
	}

	parsedTime, err := time.ParseInLocation("20060102", icalValue, time.Local)
	if err != nil {
		return LocalTime{}, fmt.Errorf("failed to parse date '%s': %w", icalValue, err)
	}

	return LocalTime{parsedTime}, nil
}

// GetEventStartTime extracts start time from an ical.Event as LocalTime
func GetEventStartTime(event *ical.Event) (LocalTime, error) {
	if event == nil {
		return LocalTime{}, fmt.Errorf("event is nil")
	}

	dtStartProp := event.Props.Get("DTSTART")
	if dtStartProp == nil {
		return LocalTime{}, fmt.Errorf("event has no DTSTART property")
	}

	// Check if it's a date-only property (all-day event)
	if valueParam := dtStartProp.Params.Get("VALUE"); valueParam == "DATE" {
		return ParseICalDate(dtStartProp.Value)
	}

	return ParseICalDateTime(dtStartProp.Value)
}

// GetEventEndTime extracts end time from an ical.Event as LocalTime
func GetEventEndTime(event *ical.Event) (LocalTime, error) {
	if event == nil {
		return LocalTime{}, fmt.Errorf("event is nil")
	}

	dtEndProp := event.Props.Get("DTEND")
	if dtEndProp == nil {
		// If no DTEND, try DURATION
		durationProp := event.Props.Get("DURATION")
		if durationProp != nil {
			startTime, err := GetEventStartTime(event)
			if err != nil {
				return LocalTime{}, fmt.Errorf("failed to get start time for duration calculation: %w", err)
			}
			
			duration, err := time.ParseDuration(durationProp.Value)
			if err != nil {
				return LocalTime{}, fmt.Errorf("failed to parse duration '%s': %w", durationProp.Value, err)
			}
			
			return startTime.Add(duration), nil
		}
		
		return LocalTime{}, fmt.Errorf("event has no DTEND or DURATION property")
	}

	// Check if it's a date-only property (all-day event)
	if valueParam := dtEndProp.Params.Get("VALUE"); valueParam == "DATE" {
		return ParseICalDate(dtEndProp.Value)
	}

	return ParseICalDateTime(dtEndProp.Value)
}

// GetEventTimes extracts both start and end times from an ical.Event as LocalTime
func GetEventTimes(event *ical.Event) (LocalTime, LocalTime, error) {
	startTime, err := GetEventStartTime(event)
	if err != nil {
		return LocalTime{}, LocalTime{}, fmt.Errorf("failed to get start time: %w", err)
	}

	endTime, err := GetEventEndTime(event)
	if err != nil {
		return LocalTime{}, LocalTime{}, fmt.Errorf("failed to get end time: %w", err)
	}

	return startTime, endTime, nil
}

// PreserveTimezoneFormat formats a LocalTime for iCal storage while preserving the original format
// This prevents timezone corruption when updating events
func PreserveTimezoneFormat(localTime LocalTime, originalValue string) string {
	if originalValue == "" {
		// No original value, default to local format (no timezone)
		return localTime.Time.Format("20060102T150405")
	}

	if strings.HasSuffix(originalValue, "Z") {
		// Original was UTC, format as UTC
		return localTime.Time.UTC().Format("20060102T150405Z")
	}

	// Original was local time (no timezone suffix), keep local
	return localTime.Time.Format("20060102T150405")
}