package util

import (
	"time"

	"github.com/samuelstranges/chronos/timezone"
	types "github.com/samuelstranges/chronos/types"
)

// CreateMidnightDate creates a time.Time at midnight (00:00:00) for the given date
func CreateMidnightDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// CreateEndOfDayDate creates a time.Time at end of day (23:59:59) for the given date
func CreateEndOfDayDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

// GetWeekStartingOn calculates the start of the week containing the given time
func GetWeekStartingOn(t time.Time, startDay time.Weekday) time.Time {
	weekday := t.Weekday()
	daysBack := (int(weekday) - int(startDay) + types.DaysPerWeek) % types.DaysPerWeek
	selectedDate := t.AddDate(0, 0, -daysBack)
	// Return start of day (midnight)
	return CreateMidnightDate(selectedDate)
}

// CellToTime converts a cell index to the corresponding time
func CellToTime(cell int, zoomLevel types.ZoomLevel, weekStart time.Time, day int) time.Time {
	// Current time in mins
	currentTimeFromMidnightInMins := cell * int(zoomLevel)

	dayDate := weekStart.AddDate(0, 0, day)

	// Convert to a rough time.Time
	hours := currentTimeFromMidnightInMins / types.MinsPerHour
	minutesLeftOver := currentTimeFromMidnightInMins % types.MinsPerHour
	return time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), hours, minutesLeftOver, 0, 0, weekStart.Location())
}

func TimeToCell(currentTime time.Time, zoomLevel types.ZoomLevel) int {
	hoursFromMidnight := currentTime.Hour()
	minutesFromHour := currentTime.Minute()

	totalMinutesFromMidnight := (hoursFromMidnight * types.MinsPerHour) + minutesFromHour

	newCell := totalMinutesFromMidnight / int(zoomLevel)

	return newCell
}

func GetCellsPerDay(zoom types.ZoomLevel) int {
	minsPerDay := types.HoursPerDay * types.MinsPerHour
	return minsPerDay / int(zoom)
}

// GetDayOfWeek calculates which day of the week (0-6) an event time falls on relative to week start
func GetDayOfWeek(eventTime time.Time, weekStart time.Time) time.Weekday {
	// FIX: Use Date() to extract date components instead of Hours() division
	// The direct time subtraction was causing timezone-related precision issues
	eventYear, eventMonth, eventDay := eventTime.Date()
	weekYear, weekMonth, weekDay := weekStart.Date()

	eventDateOnly := time.Date(eventYear, eventMonth, eventDay, 0, 0, 0, 0, time.Local)
	weekStartOnly := time.Date(weekYear, weekMonth, weekDay, 0, 0, 0, 0, time.Local)

	// Calculate days using proper date arithmetic (no division/hours)
	daysSinceStart := 0
	if eventDateOnly.After(weekStartOnly) {
		testDate := weekStartOnly
		for testDate.Before(eventDateOnly) {
			testDate = testDate.AddDate(0, 0, 1)
			daysSinceStart++
		}
	} else if eventDateOnly.Before(weekStartOnly) {
		testDate := weekStartOnly
		for testDate.After(eventDateOnly) {
			testDate = testDate.AddDate(0, 0, -1)
			daysSinceStart--
		}
	}

	// Clamp to valid week range [0-6]
	return time.Weekday(max(0, min(daysSinceStart, types.MaxDayIndex)))
}

// GetDayOfWeekFromLocalTime calculates which day of the week a LocalTime falls on relative to week start
func GetDayOfWeekFromLocalTime(eventTime timezone.LocalTime, weekStart time.Time) time.Weekday {
	return GetDayOfWeek(eventTime.Time, weekStart)
}
