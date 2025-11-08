package week_view_all_day_grid

import (
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// CalculateAllDayEventGridFromInstances creates a grid of all-day events directly from EventInstances
// NOTE: eventInstances should already be filtered to contain only all-day events
func CalculateAllDayEventGridFromInstances(eventInstances []*types.EventInstance, weekStart time.Time) (types.AllDayEventGrid, []int, error) {
	// Note: eventInstances already come sorted and filtered from EventManager
	// Calculate accurate counts per day using computed times
	counts := make([]int, types.DaysPerWeek)
	for day := range types.DaysPerWeek {
		dayDate := weekStart.AddDate(0, 0, day)
		for _, instance := range eventInstances {
			occurs, err := util.AllDayEventOccursOnDay(instance.OriginalEvent, dayDate)
			if err == nil && occurs {
				counts[day]++
			}
		}
	}

	// Convert to spanning events - this will need to be updated to work with instances
	spanningEvents := convertInstancesToAllDaySpanningEvents(eventInstances, weekStart)

	// Assign to rows to avoid overlaps
	eventGrid := types.AllDayEventGrid{
		StartDate: weekStart,
		EventRows: [types.AllDayCells][]types.AllDaySpanningEvent{},
	}

	assignEventsToAllDayRows(spanningEvents, &eventGrid)

	return eventGrid, counts, nil
}


// convertInstancesToAllDaySpanningEvents converts EventInstances to AllDaySpanningEvents
func convertInstancesToAllDaySpanningEvents(instances []*types.EventInstance, weekStart time.Time) []types.AllDaySpanningEvent {
	var spanningEvents []types.AllDaySpanningEvent

	for _, instance := range instances {
		// Use existing overlap utility to include events that overlap with the current week
		// This correctly handles multi-day events that span across week boundaries
		weekEnd := weekStart.AddDate(0, 0, types.DaysPerWeek)

		// Skip events that don't overlap with this week using existing RFC 5545 compliant logic
		if !util.EventInstanceOverlapsTimeRange(instance, weekStart, weekEnd) {
			continue
		}

		// Get proper event color with fallback hierarchy (event color -> calendar color -> default)
		calendarColor := util.GetCalendarColor(instance.Calendar)
		eventColor := util.GetEventColor(instance.OriginalEvent, calendarColor)

		// Create spanning event using computed times
		spanningEvent := types.AllDaySpanningEvent{
			Event:         instance.OriginalEvent, // Still need the original event for display
			EventInstance: instance,               // Store the instance for computed times
			StartDay:      int(util.GetDayOfWeekFromLocalTime(instance.ComputedStart, weekStart)),
			EndDay:        int(util.GetDayOfWeekFromLocalTime(instance.ComputedEnd.AddDate(0, 0, -1), weekStart)), // Subtract 1 day since DTEND is non-inclusive
			CalendarColor: eventColor,                                                                // Use event color instead of just calendar color
		}

		spanningEvents = append(spanningEvents, spanningEvent)
	}

	return spanningEvents
}

