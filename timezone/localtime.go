package timezone

import (
	"fmt"
	"strings"
	"time"
)

// LocalTime is always in user's local timezone to prevent timezone bugs
// This type enforces that all calendar display operations use local time
type LocalTime struct {
	time.Time
}

// NewLocalTime ensures time is in local timezone
func NewLocalTime(t time.Time) LocalTime {
	return LocalTime{t.In(time.Local)}
}

// NewLocalTimeFromDateTime creates LocalTime from separate date and time strings
func NewLocalTimeFromDateTime(dateStr, timeStr string) (LocalTime, error) {
	dateTimeStr := dateStr + " " + timeStr
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, time.Local)
	if err != nil {
		return LocalTime{}, fmt.Errorf("failed to parse date/time '%s': %w", dateTimeStr, err)
	}
	return LocalTime{parsedTime}, nil
}

// Add enforces local timezone on any arithmetic
func (lt LocalTime) Add(d time.Duration) LocalTime {
	return NewLocalTime(lt.Time.Add(d))
}

// AddDate adds years, months, and days to the LocalTime
func (lt LocalTime) AddDate(years, months, days int) LocalTime {
	return NewLocalTime(lt.Time.AddDate(years, months, days))
}

// Date returns the year, month, and day of the LocalTime
func (lt LocalTime) Date() (year int, month time.Month, day int) {
	return lt.Time.Date()
}

// Sub returns duration between two LocalTimes
func (lt LocalTime) Sub(other LocalTime) time.Duration {
	return lt.Time.Sub(other.Time)
}

// Before checks if this time is before another LocalTime
func (lt LocalTime) Before(other LocalTime) bool {
	return lt.Time.Before(other.Time)
}

// After checks if this time is after another LocalTime
func (lt LocalTime) After(other LocalTime) bool {
	return lt.Time.After(other.Time)
}

// Equal checks if this time equals another LocalTime
func (lt LocalTime) Equal(other LocalTime) bool {
	return lt.Time.Equal(other.Time)
}

// Format always formats the time in local timezone
func (lt LocalTime) Format(layout string) string {
	return lt.Time.Format(layout)
}

// ToUTC converts the local time to UTC for storage in iCal format
func (lt LocalTime) ToUTC() time.Time {
	return lt.Time.UTC()
}

// ToICalFormat formats time for iCal storage, preserving the original timezone context
func (lt LocalTime) ToICalFormat(originalValue string) string {
	if originalValue == "" || !strings.HasSuffix(originalValue, "Z") {
		// Original was local time (no timezone suffix), keep local
		return lt.Time.Format("20060102T150405")
	}
	// Original was UTC, format as UTC
	return lt.Time.UTC().Format("20060102T150405Z")
}

// String returns a readable representation
func (lt LocalTime) String() string {
	return fmt.Sprintf("LocalTime[%s]", lt.Time.Format("2006-01-02 15:04:05 MST"))
}