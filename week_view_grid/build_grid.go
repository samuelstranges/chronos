package week_view_grid

import (
	"fmt"
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/util"
)

// GRID STRUCTURE & ORCHESTRATION
//
// This file handles everything about:
// 1. TIME CALCULATIONS: Converting times to cell positions, handling day boundaries
// 2. GRID STRUCTURE: Initializing and managing the overall week/day/cell data structure
// 3. ORCHESTRATION: Coordinating the entire pipeline from EventInstances → WeekEventGrid
//
// Key responsibilities:
// - Time-to-cell calculations and day boundary handling
// - Grid initialization (WeekEventGrid, DayLayout structures)
// - Overall flow coordination (calling process_events_for_cells.go and build_cell.go)
// - Range operations (placing events across multiple cells/days)
//
// Uses: build_cell.go (for column assignment and cell operations)

func clipEventTimesToDayBoundaries(eventStartTime time.Time, eventEndTime time.Time, day int, weekStart time.Time) (clippedStart, clippedEnd time.Time) {
	// Useful so events dont render after midnight
	dayStart := weekStart.AddDate(0, 0, day)
	dayEnd := dayStart.AddDate(0, 0, 1) // Next day at midnight
	// Clip event times to day's boundaries
	clippedStart = eventStartTime
	if eventStartTime.Before(dayStart) {
		clippedStart = dayStart
	}

	clippedEnd = eventEndTime
	if eventEndTime.After(dayEnd) {
		clippedEnd = dayEnd
	}

	return clippedStart, clippedEnd
}

// clipEventTimesToDayBoundariesLocalTime clips LocalTime event times to day boundaries
func clipEventTimesToDayBoundariesLocalTime(eventStartTime timezone.LocalTime, eventEndTime timezone.LocalTime, day int, weekStart time.Time) (clippedStart, clippedEnd time.Time) {
	return clipEventTimesToDayBoundaries(eventStartTime.Time, eventEndTime.Time, day, weekStart)
}

// getCellRangeForEventInstanceWithinDay calculates cell range using computed times from EventInstance
func getCellRangeForEventInstanceWithinDay(instance *types.EventInstance, weekStart time.Time, zoomLevel types.ZoomLevel, day int) (startCell, endCell int, eventInDay bool) {
	// Use computed times from EventInstance instead of parsing from iCal
	eventStartTime := instance.ComputedStart
	eventEndTime := instance.ComputedEnd

	clippedStart, clippedEnd := clipEventTimesToDayBoundariesLocalTime(eventStartTime, eventEndTime, day, weekStart)

	// Check if the event actually spans into this day after clipping
	if !clippedEnd.After(clippedStart) {
		return 0, 0, false // Event doesn't actually overlap with this day
	}

	startCell = calculateStartCell(clippedStart, zoomLevel)
	endCell = CalculateEndCell(clippedEnd, zoomLevel)

	return startCell, endCell, true
}

// getMaxColumnInGroup finds the highest column number in a collision group
func getMaxColumnInGroup(group types.CollisionGroup) int {
	maxColumn := 0
	for _, event := range group {
		if event.Column > maxColumn {
			maxColumn = event.Column
		}
	}
	return maxColumn
}

// placeEventInWeekGrid places a single event across all days it spans
func placeEventInWeekGrid(weekLayout *types.WeekEventGrid, posEvent types.PositionedEvent, group types.CollisionGroup, weekStart time.Time, zoomLevel types.ZoomLevel) {
	maxColumn := getMaxColumnInGroup(group)
	requiredColumns := maxColumn + 1

	for day := range types.DaysPerWeek {
		startCell, endCell, eventInDay := getCellRangeForEventInstanceWithinDay(posEvent.Instance, weekStart, zoomLevel, day)
		if eventInDay {
			// Ensure all cells in the range can hold the required columns
			for cell := startCell; cell <= endCell; cell++ {
				ensureCellLayoutCapacity(&weekLayout.DayLayouts[day], cell, requiredColumns)
			}
			// Place the event in all cells in the range
			for cell := startCell; cell <= endCell; cell++ {
				placeEventInSingleCell(&weekLayout.DayLayouts[day], posEvent, cell, cell == startCell, cell == endCell)
			}
		}
	}
}

func buildWeekEventGridFromGroups(collisionGroups []types.CollisionGroup, weekStart time.Time, zoomLevel types.ZoomLevel) types.WeekEventGrid {
	weekLayout := types.WeekEventGrid{
		StartDate: weekStart,
	}

	// Initialize empty day layouts
	for i := range types.DaysPerWeek {
		weekLayout.DayLayouts[i] = types.DayLayout{
			Date:        weekStart.AddDate(0, 0, i),
			CellLayouts: make(map[int]types.CellLayout),
		}
	}

	// Place all events from all collision groups into the grid
	for _, group := range collisionGroups {
		for _, posEvent := range group {
			placeEventInWeekGrid(&weekLayout, posEvent, group, weekStart, zoomLevel)
		}
	}

	return weekLayout
}

// preparePositionedEvents converts EventInstances to PositionedEvents with calendar colors
func preparePositionedEvents(eventInstances []*types.EventInstance) []types.PositionedEvent {
	var allEventsInWeek []types.PositionedEvent
	
	for _, instance := range eventInstances {
		calendarColor := util.GetCalendarColor(instance.Calendar)
		posEvent := types.PositionedEvent{
			Instance:      instance,
			Column:        unassigned,
			CalendarColor: calendarColor,
		}
		allEventsInWeek = append(allEventsInWeek, posEvent)
	}
	
	return allEventsInWeek
}


// BuildTimedGrid builds a week event grid directly from EventInstances
// This is the new approach that works with both regular and recurring events
func BuildTimedGrid(eventManager types.EventManagerInterface, weekModel *types.WeekModel, zoomLevel types.ZoomLevel) (types.WeekEventGrid, error) {
	weekStart := weekModel.CurrentlyViewedWeek

	// Step 1: Get event instances from EventManager
	eventInstances, err := eventManager.GetTimedEventsForWeek(weekModel)
	if err != nil {
		return types.WeekEventGrid{StartDate: weekStart}, fmt.Errorf("failed to get event instances: %w", err)
	}

	// Step 2: Convert to PositionedEvents with calendar colors
	sortedEvents := preparePositionedEvents(eventInstances)

	// Step 3: Find and sort collision groups
	collisionGroups, err := processCollisionGroups(sortedEvents, weekStart, zoomLevel)
	if err != nil {
		return types.WeekEventGrid{}, err
	}

	// Step 4: Create independent copies to prevent cross-zoom contamination
	independentGroups := createIndependentGroups(collisionGroups)

	// Step 5: Assign columns to all groups
	processedGroups, err := assignColumnsToGroups(independentGroups, weekStart, zoomLevel)
	if err != nil {
		return types.WeekEventGrid{}, err
	}

	// Step 6: Build final grid structure
	weekLayout := buildWeekEventGridFromGroups(processedGroups, weekStart, zoomLevel)

	return weekLayout, nil
}
