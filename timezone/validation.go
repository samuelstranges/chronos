package timezone

import (
	"fmt"
	"strings"
	"time"
)

// ValidateTimeRange ensures end time is after start time
func ValidateTimeRange(startTime, endTime LocalTime) error {
	if !endTime.After(startTime) {
		return fmt.Errorf("end time (%s) must be after start time (%s)", 
			endTime.Format("2006-01-02 15:04"), 
			startTime.Format("2006-01-02 15:04"))
	}
	return nil
}

// IsValidICalDateTime validates an iCal datetime string format
func IsValidICalDateTime(value string) bool {
	if value == "" {
		return false
	}

	// Check UTC format (20060102T150405Z)
	if strings.HasSuffix(value, "Z") {
		_, err := time.Parse("20060102T150405Z", value)
		return err == nil
	}

	// Check local format (20060102T150405)
	_, err := time.Parse("20060102T150405", value)
	return err == nil
}

// IsValidICalDate validates an iCal date string format (for all-day events)
func IsValidICalDate(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.Parse("20060102", value)
	return err == nil
}

// IsValidStartTime validates a start time string in any accepted format
func IsValidStartTime(value string) bool {
	if value == "" {
		return false
	}
	return IsValidICalDateTime(value) || IsValidICalDate(value)
}

// IsValidEndTime validates an end time string in any accepted format
func IsValidEndTime(value string) bool {
	if value == "" {
		return false
	}
	return IsValidICalDateTime(value) || IsValidICalDate(value)
}

// ValidateICalTimeRange validates that two iCal time strings form a valid range
func ValidateICalTimeRange(startValue, endValue string) error {
	if startValue == "" || endValue == "" {
		return nil // Skip validation if either is empty
	}

	var startTime, endTime time.Time
	var err error

	// Parse start time
	if IsValidICalDate(startValue) {
		startTime, err = time.Parse("20060102", startValue)
	} else if strings.HasSuffix(startValue, "Z") {
		startTime, err = time.Parse("20060102T150405Z", startValue)
	} else {
		startTime, err = time.ParseInLocation("20060102T150405", startValue, time.Local)
	}
	if err != nil {
		return fmt.Errorf("invalid start time format: %s", startValue)
	}

	// Parse end time
	if IsValidICalDate(endValue) {
		endTime, err = time.Parse("20060102", endValue)
	} else if strings.HasSuffix(endValue, "Z") {
		endTime, err = time.Parse("20060102T150405Z", endValue)
	} else {
		endTime, err = time.ParseInLocation("20060102T150405", endValue, time.Local)
	}
	if err != nil {
		return fmt.Errorf("invalid end time format: %s", endValue)
	}

	// Validate range
	if !endTime.After(startTime) {
		return fmt.Errorf("end time must be after start time")
	}

	return nil
}