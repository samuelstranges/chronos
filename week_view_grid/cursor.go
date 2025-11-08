package week_view_grid

import (
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// GetEventUnderCursor returns the event at the current cursor position, or nil if none
func GetEventUnderCursor(weekModel *types.WeekModel) *types.PositionedEvent {
	return getEventAtPosition(weekModel, weekModel.Cursor)
}

// CellToTimeAtCursor returns the time at the current cursor position
func CellToTimeAtCursor(weekModel types.WeekModel) time.Time {
	return util.CellToTime(weekModel.Cursor.Cell, weekModel.CurrentZoom, weekModel.CurrentlyViewedWeek, weekModel.Cursor.Day)
}

// ClampCursorColumn ensures cursor column is valid for the current cell
func ClampCursorColumn(weekModel *types.WeekModel, cursor types.CursorPositionInWeek) types.CursorPositionInWeek {
	maxColumns := GetCellColumnCount(weekModel, cursor.Day, cursor.Cell)
	cursor.EventColumn = max(0, min(cursor.EventColumn, maxColumns-1)) // clamp
	return cursor
}


