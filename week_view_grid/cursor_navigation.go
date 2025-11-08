package week_view_grid

import (
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// JumpToTime positions cursor at the specified time
func JumpToTime(weekModel *types.WeekModel, targetTime time.Time) {
	if targetTime.IsZero() {
		return
	}

	// Always set to the week containing the target time (no-op if already correct)
	weekModel.CurrentlyViewedWeek = util.GetWeekStartingOn(targetTime, time.Sunday)

	// Position cursor at the target time
	day := int(util.GetDayOfWeek(targetTime, weekModel.CurrentlyViewedWeek))
	cell := util.TimeToCell(targetTime, weekModel.CurrentZoom)

	weekModel.Cursor = types.CursorPositionInWeek{
		Day:         day,
		Cell:        cell,
		EventColumn: 0,
	}
}

// JumpToCurrentTime positions cursor at current time
func JumpToCurrentTime(weekModel *types.WeekModel) {
	JumpToTime(weekModel, time.Now())
}

// JumpToStartOfDay positions cursor at start of current day
func JumpToStartOfDay(weekModel *types.WeekModel) {
	currentTime := CellToTimeAtCursor(*weekModel)
	startOfDay := util.CreateMidnightDate(currentTime)
	JumpToTime(weekModel, startOfDay)
}

// JumpToEndOfDay positions cursor at end of current day
func JumpToEndOfDay(weekModel *types.WeekModel) {
	currentTime := CellToTimeAtCursor(*weekModel)
	endOfDay := util.CreateEndOfDayDate(currentTime)
	JumpToTime(weekModel, endOfDay)
}

// JumpToNextWeek moves to same time in next week
func JumpToNextWeek(weekModel *types.WeekModel) {
	currentTime := CellToTimeAtCursor(*weekModel)
	nextWeekTime := currentTime.AddDate(0, 0, 7)
	JumpToTime(weekModel, nextWeekTime)
}

// JumpToPreviousWeek moves to same time in previous week
func JumpToPreviousWeek(weekModel *types.WeekModel) {
	currentTime := CellToTimeAtCursor(*weekModel)
	prevWeekTime := currentTime.AddDate(0, 0, -7)
	JumpToTime(weekModel, prevWeekTime)
}

// FindFirstEventOfDay searches for first event in current day using grid navigation
// Returns true if event found and cursor positioned
func FindFirstEventOfDay(weekModel *types.WeekModel) bool {
	zoomIndex := GetZoomIndex(weekModel.CurrentZoom)
	if zoomIndex == -1 {
		return false
	}

	currentDay := weekModel.Cursor.Day
	dayLayout := weekModel.WeekEventGrids[zoomIndex].DayLayouts[currentDay]

	// Get all cell keys and search them in order
	maxCells := util.GetCellsPerDay(weekModel.CurrentZoom)
	for cell := 0; cell < maxCells; cell++ {
		cellLayout, exists := dayLayout.CellLayouts[cell]
		if !exists {
			continue
		}

		for col := 0; col < len(cellLayout); col++ {
			if cellLayout[col].Instance != nil {
				weekModel.Cursor.Day = currentDay
				weekModel.Cursor.Cell = cell
				weekModel.Cursor.EventColumn = col
				return true
			}
		}
	}

	return false
}

// FindLastEventOfDay searches for last event in current day using grid navigation
// Returns true if event found and cursor positioned
func FindLastEventOfDay(weekModel *types.WeekModel) bool {
	zoomIndex := GetZoomIndex(weekModel.CurrentZoom)
	if zoomIndex == -1 {
		return false
	}

	currentDay := weekModel.Cursor.Day
	dayLayout := weekModel.WeekEventGrids[zoomIndex].DayLayouts[currentDay]

	// Search from end of day backward
	maxCells := util.GetCellsPerDay(weekModel.CurrentZoom)
	for cell := maxCells - 1; cell >= 0; cell-- {
		cellLayout, exists := dayLayout.CellLayouts[cell]
		if !exists {
			continue
		}

		for col := len(cellLayout) - 1; col >= 0; col-- {
			if cellLayout[col].Instance != nil {
				weekModel.Cursor.Day = currentDay
				weekModel.Cursor.Cell = cell
				weekModel.Cursor.EventColumn = col
				return true
			}
		}
	}

	return false
}
