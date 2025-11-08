package week_view_grid

import (
	"github.com/samuelstranges/chronos/types"
)

// InvalidCursorPosition represents an invalid cursor position
var InvalidCursorPosition = types.CursorPositionInWeek{Day: -1, Cell: -1, EventColumn: -1}

// getEventAtPosition returns the event at the specified cursor position, or nil if none
func getEventAtPosition(weekModel *types.WeekModel, cursor types.CursorPositionInWeek) *types.PositionedEvent {
	if cursor.Day < 0 || cursor.Day >= 7 || cursor.Cell < 0 {
		return nil
	}

	// Ensure zoom index valid
	zoomIndex := GetZoomIndex(weekModel.CurrentZoom)
	if zoomIndex == -1 {
		return nil
	}

	if cursor.Day >= len(weekModel.WeekEventGrids[zoomIndex].DayLayouts) {
		return nil
	}

	dayLayout := weekModel.WeekEventGrids[zoomIndex].DayLayouts[cursor.Day]
	cellLayout, exists := dayLayout.CellLayouts[cursor.Cell]
	if !exists {
		return nil
	}

	if cursor.EventColumn < 0 || cursor.EventColumn >= len(cellLayout) {
		if len(cellLayout) > 0 {
			return &cellLayout[0] // TODO: This fallback logic is questionable
		}
		return nil
	}

	return &cellLayout[cursor.EventColumn]
}

// GetCellColumnCount returns the number of event columns in a specific cell
func GetCellColumnCount(weekModel *types.WeekModel, day, cell int) int {
	zoomIndex := GetZoomIndex(weekModel.CurrentZoom)
	if zoomIndex == -1 || day < 0 || day >= 7 || cell < 0 {
		return 0
	}

	dayLayout := weekModel.WeekEventGrids[zoomIndex].DayLayouts[day]
	cellLayout, exists := dayLayout.CellLayouts[cell]
	if !exists {
		return 0
	}

	return len(cellLayout)
}
