package week_view_all_day_grid

import (
	"sort"

	"github.com/samuelstranges/chronos/types"
)

// allDayEventsOverlap checks if two all-day events overlap in their day span
func allDayEventsOverlap(a, b types.AllDaySpanningEvent) bool {
	return a.EndDay >= b.StartDay && b.EndDay >= a.StartDay
}

// assignEventsToAllDayRows assigns events to rows avoiding overlaps
func assignEventsToAllDayRows(events []types.AllDaySpanningEvent, grid *types.AllDayEventGrid) {
	// Sort events by start day, then by length (longer events first)
	sort.Slice(events, func(i, j int) bool {
		if events[i].StartDay != events[j].StartDay {
			return events[i].StartDay < events[j].StartDay
		}
		// Longer events get priority
		iLength := events[i].EndDay - events[i].StartDay + 1
		jLength := events[j].EndDay - events[j].StartDay + 1
		return iLength > jLength
	})

	// Track overflow
	overflowCount := 0

	// Assign events to rows
	for _, event := range events {
		assigned := false
		for row := 0; row < types.AllDayCells && !assigned; row++ {
			// Check if this row has space for the event
			canFit := true
			for _, existingEvent := range grid.EventRows[row] {
				if allDayEventsOverlap(event, existingEvent) {
					canFit = false
					break
				}
			}

			if canFit {
				grid.EventRows[row] = append(grid.EventRows[row], event)
				assigned = true
			}
		}

		// Track events that couldn't be assigned
		if !assigned {
			overflowCount++
		}
	}

	// Set overflow indicators
	grid.OverflowCount = overflowCount
	grid.HasOverflow = overflowCount > 0
}