package week_view_grid

// Functions relating to zooming in and out on the week view

import (
	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

var zoomTransitions = map[string]map[types.ZoomLevel]types.ZoomLevel{
	"in": {
		types.Zoom30Min: types.Zoom15Min,
		types.Zoom15Min: types.Zoom5Min,
		types.Zoom5Min:  types.Zoom1Min,
		types.Zoom1Min:  types.Zoom1Min, // already at max
	},
	"out": {
		types.Zoom1Min:  types.Zoom5Min,
		types.Zoom5Min:  types.Zoom15Min,
		types.Zoom15Min: types.Zoom30Min,
		types.Zoom30Min: types.Zoom30Min, // already at min
	},
}

func zoom(m types.WeekModel, direction string) types.WeekModel {
	newZoom := zoomTransitions[direction][m.CurrentZoom]

	// Convert cursor to new zoom level
	oldTime := util.CellToTime(m.Cursor.Cell, m.CurrentZoom, m.CurrentlyViewedWeek, 0)
	newCell := util.TimeToCell(oldTime, newZoom)

	m.Cursor.Cell = min(newCell, util.GetCellsPerDay(newZoom)-1)
	m.CurrentZoom = newZoom

	return m
}

func ZoomIn(m types.WeekModel) types.WeekModel  { return zoom(m, "in") }
func ZoomOut(m types.WeekModel) types.WeekModel { return zoom(m, "out") }

// GetZoomIndex returns the array index for a given zoom level
func GetZoomIndex(zoom types.ZoomLevel) int {
	for i, level := range types.ZoomLevels {
		if level == zoom {
			return i
		}
	}
	return -1
}
