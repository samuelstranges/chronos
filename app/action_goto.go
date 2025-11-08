package app

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// NAVIGATION COMMAND IMPLEMENTATIONS

// GoToHour jumps cursor to the specified hour on the current day
func GoToHour(weekModel *types.WeekModel, hour int) error {
	if hour < 0 || hour > 23 {
		return fmt.Errorf("invalid hour: %d (must be 00-23)", hour)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Create target time at specified hour, keeping current day and timezone
	targetTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		hour, 0, 0, 0, currentTime.Location())

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// GoToMinute jumps cursor to the specified minute within the current hour
func GoToMinute(weekModel *types.WeekModel, minute int) error {
	if minute < 0 || minute > 59 {
		return fmt.Errorf("invalid minute: %d (must be 00-59)", minute)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Create target time with same hour but specified minute
	targetTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), minute, 0, 0, currentTime.Location())

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// GoToSpecificTime jumps cursor to the specified hour and minute on current day
func GoToSpecificTime(weekModel *types.WeekModel, hour, minute int) error {
	if hour < 0 || hour > 23 {
		return fmt.Errorf("invalid hour: %d (must be 00-23)", hour)
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("invalid minute: %d (must be 00-59)", minute)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Create target time at specified hour and minute, keeping current day and timezone
	targetTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
		hour, minute, 0, 0, currentTime.Location())

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// GoToMonthDay jumps to the specified day of the month, preserving time
func GoToMonthDay(weekModel *types.WeekModel, dayOfMonth int) error {
	if dayOfMonth < 1 || dayOfMonth > 31 {
		return fmt.Errorf("invalid day of month: %d (must be 01-31)", dayOfMonth)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Try to create target date with specified day of month
	targetTime := time.Date(currentTime.Year(), currentTime.Month(), dayOfMonth,
		currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())

	// Check if the day actually exists in this month
	if targetTime.Day() != dayOfMonth {
		return fmt.Errorf("day %02d does not exist in %s %d", dayOfMonth,
			currentTime.Month().String(), currentTime.Year())
	}

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// GoToMonth jumps to the specified month, preserving day and time if possible
func GoToMonth(weekModel *types.WeekModel, month int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %02d (must be 01-12)", month)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Try to create target date with specified month
	targetTime := time.Date(currentTime.Year(), time.Month(month), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())

	// Check if the day exists in the target month (handles Feb 29 -> Feb 28, etc)
	if targetTime.Month() != time.Month(month) {
		// Day doesn't exist in target month, use last day of target month
		lastDayOfMonth := time.Date(currentTime.Year(), time.Month(month+1), 0,
			currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())
		targetTime = lastDayOfMonth
	}

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// GoToYear jumps to the specified year, preserving month, day and time if possible
func GoToYear(weekModel *types.WeekModel, year int) error {
	if year < 1900 || year > 2100 {
		return fmt.Errorf("invalid year: %d (must be 1900-2100)", year)
	}

	// Get current time from cursor position
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)

	// Try to create target date with specified year
	targetTime := time.Date(year, currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())

	// Handle leap year edge case (Feb 29 -> Feb 28)
	if targetTime.Month() != currentTime.Month() {
		// Day doesn't exist in target year (probably Feb 29 -> Feb 28)
		targetTime = time.Date(year, currentTime.Month(), 28,
			currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())
	}

	week_view_grid.JumpToTime(weekModel, targetTime)
	return nil
}

// NAVIGATION PATTERN HANDLERS
// These handlers call the business logic functions in commands/pattern_handlers.go

// HandleGoToHour jumps cursor to the specified hour on the current day (gh##)
func HandleGoToHour(m Model, hour int) (tea.Model, tea.Cmd) {
	if err := GoToHour(&m.WeekModel, hour); err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to go to hour")
	}
	return m, nil
}

// HandleGoToMinute jumps cursor to the specified minute within the current hour (gm##)
func HandleGoToMinute(m Model, minute int) (tea.Model, tea.Cmd) {
	if err := GoToMinute(&m.WeekModel, minute); err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to go to minute")
	}
	return m, nil
}

// HandleGoToTime jumps cursor to the specified hour and minute on current day (gt####)
func HandleGoToTime(m Model, time int) (tea.Model, tea.Cmd) {
	// Convert to 4-digit string with leading zeros (e.g., 130 -> "0130")
	timeStr := fmt.Sprintf("%04d", time)
	
	// Parse hour and minute from string
	hour, err1 := strconv.Atoi(timeStr[:2])
	minute, err2 := strconv.Atoi(timeStr[2:])
	
	if err1 != nil || err2 != nil {
		util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, fmt.Errorf("invalid time format"), "Failed to parse time")
		return m, nil
	}
	
	if err := GoToSpecificTime(&m.WeekModel, hour, minute); err != nil {
		util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Failed to go to time")
	}
	return m, nil
}

// HandleGoToDay jumps to the specified day of the month, preserving time (gd##)
func HandleGoToDay(m Model, day int) (tea.Model, tea.Cmd) {
	GoToMonthDay(&m.WeekModel, day)
	return m, nil
}

// HandleGoToMonth jumps to the specified month, preserving day and time if possible (gM##)
func HandleGoToMonth(m Model, month int) (tea.Model, tea.Cmd) {
	GoToMonth(&m.WeekModel, month)
	return m, nil
}

// HandleGoToYear jumps to the specified year, preserving month, day and time if possible (gY####)
func HandleGoToYear(m Model, year int) (tea.Model, tea.Cmd) {
	GoToYear(&m.WeekModel, year)
	return m, nil
}
