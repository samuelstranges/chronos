package week_view_grid

import (
	"fmt"
	"time"

	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// COLUMN ASSIGNMENT & CELL OPERATIONS
//
// This file handles everything about:
// 1. COLUMN ASSIGNMENT: Taking overlapping events and figuring out which visual column each should be in
// 2. CELL OPERATIONS: Individual cell manipulation (placing events, managing capacity)
//
// Key responsibilities:
// - Collision detection between events
// - Column assignment algorithms (greedy leftmost-fit)
// - Single cell placement and capacity management
// - All the "column fitting" logic
//
// Used by: build_grid.go (which orchestrates the overall grid assembly)

// eventsCellOverlap checks if two events overlap when rendered in cells at the given zoom level
func eventsCellOverlap(eventA, eventB types.PositionedEvent, weekStart time.Time, zoomLevel types.ZoomLevel) bool {
	// Get the cell ranges for both events across all days
	for day := range types.DaysPerWeek {
		startCellA, endCellA, eventAInDay := getCellRangeForEventInstanceWithinDay(eventA.Instance, weekStart, zoomLevel, day)
		if !eventAInDay {
			continue
		}

		startCellB, endCellB, eventBInDay := getCellRangeForEventInstanceWithinDay(eventB.Instance, weekStart, zoomLevel, day)
		if !eventBInDay {
			continue
		}

		// Check if cell ranges overlap on this day
		if util.CellRangesOverlap(startCellA, endCellA, startCellB, endCellB) {
			return true
		}
	}

	return false
}

// check if event will fit in specified column using cell-based overlap detection
func isColumnFreeForEvent(columnEvents []types.PositionedEvent, newEvent types.PositionedEvent, weekStart time.Time, zoomLevel types.ZoomLevel) (bool, error) {
	for _, existingEvent := range columnEvents {
		overlaps := eventsCellOverlap(existingEvent, newEvent, weekStart, zoomLevel)
		if overlaps {
			return false, nil
		}
	}
	return true, nil
}

// return the index of the leftmost column where the event can fit
// returns len(columns) if no column is free and needs a new one
func findAvailableColumn(event types.PositionedEvent, columns [][]types.PositionedEvent, weekStart time.Time, zoomLevel types.ZoomLevel) (int, error) {
	for columnIndex := range len(columns) {
		isFree, err := isColumnFreeForEvent(columns[columnIndex], event, weekStart, zoomLevel)
		if err != nil {
			return -1, err
		}
		if isFree {
			return columnIndex, nil
		}
	}

	// no column free... need to add a new one
	return len(columns), nil
}

// Take a collision group, and assign them columns so they don't overlap visually at the given zoom level
// uses a greedy algorithm that places events in leftmost available column
func assignColumns(collisionGroup types.CollisionGroup, weekStart time.Time, zoomLevel types.ZoomLevel) (types.CollisionGroup, error) {
	// initialize 2d column array (time and fitted events)
	columns := make([][]types.PositionedEvent, 0)

	for i, event := range collisionGroup {
		// Figure out where to place this event vertically
		columnIndex, err := findAvailableColumn(event, columns, weekStart, zoomLevel)
		if err != nil {
			return nil, err
		}
		collisionGroup[i].Column = columnIndex

		// if we need to add a new column
		for len(columns) <= columnIndex {
			columns = append(columns, make([]types.PositionedEvent, 0))
		}

		// Add event to column
		columns[columnIndex] = append(columns[columnIndex], collisionGroup[i])
	}

	// Return all events with their column assignments
	return collisionGroup, nil
}

// ensureCellLayoutCapacity ensures a cell can hold the required number of columns
func ensureCellLayoutCapacity(dayLayout *types.DayLayout, cell int, requiredColumns int) {
	if dayLayout.CellLayouts[cell] == nil {
		dayLayout.CellLayouts[cell] = make(types.CellLayout, requiredColumns)
	} else if len(dayLayout.CellLayouts[cell]) < requiredColumns {
		resized := make(types.CellLayout, requiredColumns)
		copy(resized, dayLayout.CellLayouts[cell])
		dayLayout.CellLayouts[cell] = resized
	}
}

// assignColumnsToGroups assigns columns to all collision groups
func assignColumnsToGroups(groups []types.CollisionGroup, weekStart time.Time, zoomLevel types.ZoomLevel) ([]types.CollisionGroup, error) {
	for i, group := range groups {
		assignedGroup, err := assignColumns(group, weekStart, zoomLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to assign columns to collision group %d: %w", i, err)
		}
		groups[i] = assignedGroup
	}
	return groups, nil
}

// placeEventInSingleCell places an event in one specific cell
func placeEventInSingleCell(dayLayout *types.DayLayout, posEvent types.PositionedEvent, cell int, isStartCell, isEndCell bool) {
	dayLayout.CellLayouts[cell][posEvent.Column] = types.PositionedEvent{
		Instance:      posEvent.Instance,
		Column:        posEvent.Column,
		IsStartCell:   isStartCell,
		IsEndCell:     isEndCell,
		CalendarColor: posEvent.CalendarColor,
	}
}
