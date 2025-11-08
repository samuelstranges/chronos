package navigation

import (
	"time"
	
	"github.com/samuelstranges/chronos/ical_crud"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// navigateAcrossWeeks searches for matches across weeks using time-based navigation
func navigateAcrossWeeks(weekModel *types.WeekModel, eventManager *ical_crud.EventManager, direction Direction, criteria NavigationCriteria) bool {
	currentTime := week_view_grid.CellToTimeAtCursor(*weekModel)
	
	// Find next event instance that matches our criteria
	var eventInstance *types.EventInstance
	if direction == Forward {
		eventInstance = findNextMatchingInstance(eventManager, currentTime, criteria)
	} else {
		eventInstance = findPreviousMatchingInstance(eventManager, currentTime, criteria)
	}
	
	if eventInstance == nil {
		return false
	}
	
	// Determine target time based on criteria type
	_, isEndCriteria := criteria.(EventEndCriteria)
	var targetTime time.Time
	if isEndCriteria {
		targetTime = eventInstance.ComputedEnd.Time
	} else {
		targetTime = eventInstance.ComputedStart.Time
	}
	
	// Jump to the appropriate event time
	week_view_grid.JumpToTime(weekModel, targetTime)
	return true
}

// findNextMatchingInstance finds the next event instance that matches criteria
func findNextMatchingInstance(eventManager *ical_crud.EventManager, currentTime time.Time, criteria NavigationCriteria) *types.EventInstance {
	searchRange := 365 * 24 * time.Hour // Search up to 1 year
	searchStart := currentTime
	searchEnd := currentTime.Add(searchRange)
	
	instances, err := eventManager.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return nil
	}
	
	var bestInstance *types.EventInstance
	var bestTime time.Time
	
	for _, instance := range instances {
		// Skip all-day events
		if instance.OriginalEvent != nil {
			if dtStart := instance.OriginalEvent.Props.Get("DTSTART"); dtStart != nil {
				if dtStart.Params.Get("VALUE") == "DATE" {
					continue
				}
			}
		}
		
		eventTime := instance.ComputedStart.Time
		if !eventTime.After(currentTime) {
			continue
		}
		
		// Create a positioned event to check criteria
		// Set flags based on criteria type to enable proper matching
		_, isEndCriteria := criteria.(EventEndCriteria)
		positionedEvent := types.PositionedEvent{
			Instance:    instance,
			IsStartCell: !isEndCriteria,
			IsEndCell:   isEndCriteria,
		}
		
		if !criteria.Matches(positionedEvent) {
			continue
		}
		
		if bestInstance == nil || eventTime.Before(bestTime) {
			bestInstance = instance
			bestTime = eventTime
		}
	}
	
	return bestInstance
}

// findPreviousMatchingInstance finds the previous event instance that matches criteria
func findPreviousMatchingInstance(eventManager *ical_crud.EventManager, currentTime time.Time, criteria NavigationCriteria) *types.EventInstance {
	searchRange := 365 * 24 * time.Hour // Search up to 1 year
	searchStart := currentTime.Add(-searchRange)
	searchEnd := currentTime
	
	instances, err := eventManager.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return nil
	}
	
	var bestInstance *types.EventInstance
	var bestTime time.Time
	
	for _, instance := range instances {
		// Skip all-day events
		if instance.OriginalEvent != nil {
			if dtStart := instance.OriginalEvent.Props.Get("DTSTART"); dtStart != nil {
				if dtStart.Params.Get("VALUE") == "DATE" {
					continue
				}
			}
		}
		
		eventTime := instance.ComputedStart.Time
		if !eventTime.Before(currentTime) {
			continue
		}
		
		// Create a positioned event to check criteria
		// Set flags based on criteria type to enable proper matching
		_, isEndCriteria := criteria.(EventEndCriteria)
		positionedEvent := types.PositionedEvent{
			Instance:    instance,
			IsStartCell: !isEndCriteria,
			IsEndCell:   isEndCriteria,
		}
		
		if !criteria.Matches(positionedEvent) {
			continue
		}
		
		if bestInstance == nil || eventTime.After(bestTime) {
			bestInstance = instance
			bestTime = eventTime
		}
	}
	
	return bestInstance
}