package week_view_grid

import (
	"time"

	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

func calculateStartCell(eventStartTime time.Time, zoom types.ZoomLevel) int {
	return util.TimeToCell(eventStartTime, zoom)
}

func CalculateEndCell(eventEndTime time.Time, zoom types.ZoomLevel) int {
	endCell := util.TimeToCell(eventEndTime, zoom)

	// HANDLE EDGE CASE
	// Example: Event ends at 10:00 AM with 30-min zoom
	// 10:00 AM = 600 minutes from midnight
	// 600 / 30 = cell 20
	// But the event actually ENDS at 10:00, so it occupies 9:30-10:00 (cell 19)
	// We need to check if the end time lands exactly on a cell boundary
	minutesFromMidnight := eventEndTime.Hour()*types.MinsPerHour + eventEndTime.Minute()
	landsOnCellBoundary := minutesFromMidnight%int(zoom) == 0

	// Special case: midnight (00:00) should be treated as end of previous day
	if minutesFromMidnight == 0 {
		maxCells := util.GetCellsPerDay(zoom)
		return maxCells - 1 // Last cell of the day
	}

	// Regular boundary case (not midnight)
	if landsOnCellBoundary {
		return endCell - 1
	}

	return endCell
}

