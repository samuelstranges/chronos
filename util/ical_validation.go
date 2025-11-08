package util

import "time"

// IsValidStartTime validates a DTSTART time string (supports both timed and all-day formats)
func IsValidStartTime(dtstart string) bool {
	if dtstart == "" {
		return true // Empty is valid (optional)
	}
	
	// Try datetime format first (20060102T150405Z)
	if _, err := time.Parse("20060102T150405Z", dtstart); err == nil {
		return true
	}
	
	// Try local datetime format (20060102T150405)
	if _, err := time.Parse("20060102T150405", dtstart); err == nil {
		return true
	}
	
	// Try all-day date format (20060102)
	if _, err := time.Parse("20060102", dtstart); err == nil {
		return true
	}
	
	return false
}

// IsValidEndTime validates a DTEND time string (supports both timed and all-day formats)
func IsValidEndTime(dtend string) bool {
	if dtend == "" {
		return true // Empty is valid (optional)
	}
	
	// Try datetime format first (20060102T150405Z)
	if _, err := time.Parse("20060102T150405Z", dtend); err == nil {
		return true
	}
	
	// Try local datetime format (20060102T150405)
	if _, err := time.Parse("20060102T150405", dtend); err == nil {
		return true
	}
	
	// Try all-day date format (20060102)
	if _, err := time.Parse("20060102", dtend); err == nil {
		return true
	}
	
	return false
}

// IsValidTimeRange validates that start time is before end time
func IsValidTimeRange(dtstart, dtend string) bool {
	if dtstart == "" || dtend == "" {
		return true // Can't validate if either is missing
	}
	
	// Parse start time (try all formats)
	var startTime time.Time
	var err1 error
	if startTime, err1 = time.Parse("20060102T150405Z", dtstart); err1 != nil {
		if startTime, err1 = time.Parse("20060102T150405", dtstart); err1 != nil {
			if startTime, err1 = time.Parse("20060102", dtstart); err1 != nil {
				return false // Invalid start format
			}
		}
	}
	
	// Parse end time (try all formats)
	var endTime time.Time
	var err2 error
	if endTime, err2 = time.Parse("20060102T150405Z", dtend); err2 != nil {
		if endTime, err2 = time.Parse("20060102T150405", dtend); err2 != nil {
			if endTime, err2 = time.Parse("20060102", dtend); err2 != nil {
				return false // Invalid end format
			}
		}
	}
	
	return endTime.After(startTime) // Events must have duration > 0
}

// IsValidUID validates a UID string (non-empty)
func IsValidUID(uid string) bool {
	return uid != ""
}