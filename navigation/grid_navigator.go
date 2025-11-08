package navigation

import (
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// navigateWithinWeek searches for matches within current week using grid navigation
func navigateWithinWeek(weekModel *types.WeekModel, direction Direction, criteria NavigationCriteria) bool {
	zoomIndex := week_view_grid.GetZoomIndex(weekModel.CurrentZoom)
	if zoomIndex == -1 {
		return false
	}
	
	cursor := weekModel.Cursor
	
	if direction == Forward {
		return searchGrid(weekModel, criteria, zoomIndex, cursor.Day, 7, 1,
			func(day int) int { 
				if day == cursor.Day { return cursor.Cell }
				return 0 
			},
			func(day, cell int) int {
				if day == cursor.Day && cell == cursor.Cell { return cursor.EventColumn + 1 }
				return 0
			})
	} else {
		return searchGrid(weekModel, criteria, zoomIndex, cursor.Day, -1, -1,
			func(day int) int {
				if day == cursor.Day { return cursor.Cell }
				return util.GetCellsPerDay(weekModel.CurrentZoom) - 1
			},
			func(day, cell int) int {
				if day == cursor.Day && cell == cursor.Cell {
					if cursor.EventColumn == 0 {
						return -1 // Skip this cell, move to previous cell
					}
					return cursor.EventColumn - 1
				}
				return -2 // Special value meaning "start from end of cell"
			})
	}
}

// searchGrid performs the actual grid search with parameterized iteration
func searchGrid(weekModel *types.WeekModel, criteria NavigationCriteria, zoomIndex int,
	dayStart, dayEnd, dayStep int,
	getCellStart func(int) int,
	getColStart func(int, int) int) bool {
	
	totalCells := util.GetCellsPerDay(weekModel.CurrentZoom)
	
	for day := dayStart; day != dayEnd; day += dayStep {
		dayLayout := weekModel.WeekEventGrids[zoomIndex].DayLayouts[day]
		
		cellStart := getCellStart(day)
		var cellEnd int
		if dayStep > 0 {
			cellEnd = totalCells
		} else {
			cellEnd = -1
		}
		
		for cell := cellStart; cell != cellEnd; cell += dayStep {
			cellLayout, exists := dayLayout.CellLayouts[cell]
			if !exists {
				continue
			}
			
			colStart := getColStart(day, cell)
			if colStart == -1 { // For backward search, skip current cell - we've exhausted it
				continue  
			}
			if colStart == -2 { // For backward search, start from end of cell
				colStart = len(cellLayout) - 1
			}
			
			var colEnd int
			if dayStep > 0 {
				colEnd = len(cellLayout)
			} else {
				colEnd = -1
			}
			
			for col := colStart; col != colEnd; col += dayStep {
				if col < 0 || col >= len(cellLayout) {
					continue
				}
				
				if criteria.Matches(cellLayout[col]) {
					weekModel.Cursor.Day = day
					weekModel.Cursor.Cell = cell
					weekModel.Cursor.EventColumn = col
					return true
				}
			}
		}
	}
	
	return false
}