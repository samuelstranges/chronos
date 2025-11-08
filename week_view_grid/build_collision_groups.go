package week_view_grid

import (
	"fmt"
	"sort"
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// PROCESSING FLOW STEP 1: Event Collision Detection
//
// EventInstances → CollisionGroups
//
// This file handles:
// - Finding which events overlap in time
// - Grouping overlapping events into collision groups
// - Sorting collision groups by earliest start time
// - Creating independent group copies for zoom-level isolation
//
// Next: build_cell.go assigns columns within each collision group

// Timed Event Collision Logic

const unassigned = -1

// eventOverlapsWithGroup checks if an event overlaps with any event in a collision group
func eventOverlapsWithGroup(event types.PositionedEvent, group types.CollisionGroup, weekStart time.Time, zoomLevel types.ZoomLevel) (bool, error) {
	for _, groupEvent := range group {
		overlaps := eventsCellOverlap(event, groupEvent, weekStart, zoomLevel)
		if overlaps {
			return true, nil
		}
	}
	return false, nil
}

// addEventToGroup adds an event to a collision group and marks it as processed
func addEventToGroup(event types.PositionedEvent, group *types.CollisionGroup, processed []bool, index int) {
	*group = append(*group, event)
	processed[index] = true
}

// buildCollisionGroup creates a collision group starting from a specific event
func buildCollisionGroup(startIndex int, events []types.PositionedEvent, processed []bool, weekStart time.Time, zoomLevel types.ZoomLevel) (types.CollisionGroup, error) {
	group := types.CollisionGroup{events[startIndex]}
	processed[startIndex] = true

	// Find all overlapping events
	for i := startIndex + 1; i < len(events); i++ {
		if processed[i] {
			continue
		}

		overlaps, err := eventOverlapsWithGroup(events[i], group, weekStart, zoomLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to check overlap for event %d: %w", i, err)
		}

		if overlaps {
			addEventToGroup(events[i], &group, processed, i)
		}
	}

	return group, nil
}

// getEarliestEventInGroup finds the event with the earliest start time in a collision group
func getEarliestEventInGroup(group types.CollisionGroup) types.PositionedEvent {
	if len(group) == 0 {
		return types.PositionedEvent{}
	}

	earliest := group[0]
	for _, event := range group[1:] {
		if event.Instance.ComputedStart.Before(earliest.Instance.ComputedStart) {
			earliest = event
		}
	}
	return earliest
}

// sortCollisionGroups ensures deterministic ordering of collision groups by earliest event start time
func sortCollisionGroups(groups []types.CollisionGroup) {
	sort.Slice(groups, func(i, j int) bool {
		// Empty groups go first
		if len(groups[i]) == 0 {
			return len(groups[j]) > 0
		}
		if len(groups[j]) == 0 {
			return false
		}

		// Compare by earliest start time in each group
		iEarliest := getEarliestEventInGroup(groups[i])
		jEarliest := getEarliestEventInGroup(groups[j])

		iStart := iEarliest.Instance.ComputedStart
		jStart := jEarliest.Instance.ComputedStart

		if iStart.Equal(jStart) {
			// Use title as tiebreaker for stable sorting
			iTitle := util.GetEventTitle(iEarliest.Instance.OriginalEvent)
			jTitle := util.GetEventTitle(jEarliest.Instance.OriginalEvent)
			return iTitle < jTitle
		}

		return iStart.Before(jStart)
	})
}

// findCollisionGroups groups overlapping events together
func findCollisionGroups(events []types.PositionedEvent, weekStart time.Time, zoomLevel types.ZoomLevel) ([]types.CollisionGroup, error) {
	var groups []types.CollisionGroup
	processed := make([]bool, len(events))

	for i := range events {
		if processed[i] {
			continue
		}

		group, err := buildCollisionGroup(i, events, processed, weekStart, zoomLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to build group starting at %d: %w", i, err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// processCollisionGroups finds collision groups and sorts them for deterministic ordering
func processCollisionGroups(events []types.PositionedEvent, weekStart time.Time, zoomLevel types.ZoomLevel) ([]types.CollisionGroup, error) {
	collisionGroups, err := findCollisionGroups(events, weekStart, zoomLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to find collision groups: %w", err)
	}
	
	sortCollisionGroups(collisionGroups)
	return collisionGroups, nil
}

// createIndependentGroups creates isolated copies of collision groups to prevent cross-zoom contamination
func createIndependentGroups(collisionGroups []types.CollisionGroup) []types.CollisionGroup {
	independentGroups := make([]types.CollisionGroup, len(collisionGroups))
	for i, group := range collisionGroups {
		independentGroups[i] = make(types.CollisionGroup, len(group))
		for j, event := range group {
			independentGroups[i][j] = types.PositionedEvent{
				Instance:      event.Instance,
				Column:        unassigned,
				CalendarColor: event.CalendarColor,
				IsStartCell:   event.IsStartCell,
				IsEndCell:     event.IsEndCell,
			}
		}
	}
	return independentGroups
}
