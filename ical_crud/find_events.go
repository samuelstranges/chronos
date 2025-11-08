package ical_crud

import (
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// findEventInDirection finds an EventInstance relative to the given time in the specified direction
// useEndTime determines whether to search by DTEND (true) or DTSTART (false)
// Now uses EventInstances to support recurring events and returns the actual EventInstance
func (em *EventManager) findEventInDirection(currentTime time.Time, forward bool, useEndTime bool) *types.EventInstance {
	var bestInstance *types.EventInstance
	var bestEventTime time.Time

	// Search across multiple weeks to catch events outside current view
	searchRange := types.DaysInYear * types.HoursPerDay * time.Hour // Search up to 1 year
	var searchStart, searchEnd time.Time

	if forward {
		searchStart = currentTime
		searchEnd = currentTime.Add(searchRange)
	} else {
		searchStart = currentTime.Add(-searchRange)
		searchEnd = currentTime
	}

	// Get all EventInstances in the search range (includes expanded recurring events)
	instances, err := em.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return nil
	}

	for _, instance := range instances {
		// Skip all-day events from navigation (w, b, e commands)
		if util.IsAllDayEvent(instance.OriginalEvent) {
			continue
		}

		var eventTime time.Time

		if useEndTime {
			eventTime = instance.ComputedEnd.Time
		} else {
			eventTime = instance.ComputedStart.Time
		}

		var shouldConsider bool
		var isBetter bool

		if forward { // find next event
			shouldConsider = eventTime.After(currentTime)
			isBetter = bestInstance == nil || eventTime.Before(bestEventTime)
		} else { // find previous event
			shouldConsider = eventTime.Before(currentTime)
			isBetter = bestInstance == nil || eventTime.After(bestEventTime)
		}

		if shouldConsider && isBetter {
			bestInstance = instance
			bestEventTime = eventTime
		}
	}
	return bestInstance
}

// FindNextEventInstance finds the next EventInstance after the given time across all calendars
// Returns nil if no event is found
func (em *EventManager) FindNextEventInstance(currentTime time.Time) *types.EventInstance {
	return em.findEventInDirection(currentTime, true, false)
}

// FindPreviousEventInstance finds the previous EventInstance before the given time across all calendars
// Returns nil if no event is found
func (em *EventManager) FindPreviousEventInstance(currentTime time.Time) *types.EventInstance {
	return em.findEventInDirection(currentTime, false, false)
}

// matchesSearchCriteria checks if an event matches the search field and phrase
func matchesSearchCriteria(event *ical.Event, field, phrase string) bool {
	phrase = strings.ToLower(phrase)
	field = strings.ToLower(field)

	switch field {
	case "title", "summary":
		return strings.Contains(strings.ToLower(util.GetEventTitle(event)), phrase)
	case "description":
		return strings.Contains(strings.ToLower(util.GetEventDescription(event)), phrase)
	case "location":
		return strings.Contains(strings.ToLower(util.GetEventLocation(event)), phrase)
	case "all", "":
		// Search all fields
		return strings.Contains(strings.ToLower(util.GetEventTitle(event)), phrase) ||
			strings.Contains(strings.ToLower(util.GetEventDescription(event)), phrase) ||
			strings.Contains(strings.ToLower(util.GetEventLocation(event)), phrase)
	default:
		return false
	}
}

// TODO: these 2 methods have repeated logic...

// FindNextEventInstanceMatching finds the next EventInstance after the given time that matches search criteria
func (em *EventManager) FindNextEventInstanceMatching(currentTime time.Time, field, phrase string) *types.EventInstance {
	var bestInstance *types.EventInstance
	var bestEventTime time.Time

	// Search across multiple weeks to catch events outside current view
	searchRange := types.DaysInYear * types.HoursPerDay * time.Hour // Search up to 1 year
	searchStart := currentTime
	searchEnd := currentTime.Add(searchRange)

	// Get all EventInstances in the search range
	instances, err := em.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return nil
	}

	for _, instance := range instances {
		// Skip all-day events from navigation
		if util.IsAllDayEvent(instance.OriginalEvent) {
			continue
		}

		// Check if event matches search criteria
		if !matchesSearchCriteria(instance.OriginalEvent, field, phrase) {
			continue
		}

		eventTime := instance.ComputedStart.Time
		shouldConsider := eventTime.After(currentTime)
		isBetter := bestInstance == nil || eventTime.Before(bestEventTime)

		if shouldConsider && isBetter {
			bestInstance = instance
			bestEventTime = eventTime
		}
	}
	return bestInstance
}

// FindPreviousEventInstanceMatching finds the previous EventInstance before the given time that matches search criteria
func (em *EventManager) FindPreviousEventInstanceMatching(currentTime time.Time, field, phrase string) *types.EventInstance {
	var bestInstance *types.EventInstance
	var bestEventTime time.Time

	// Search across multiple weeks to catch events outside current view
	searchRange := types.DaysInYear * types.HoursPerDay * time.Hour // Search up to 1 year
	searchStart := currentTime.Add(-searchRange)
	searchEnd := currentTime

	// Get all EventInstances in the search range
	instances, err := em.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return nil
	}

	for _, instance := range instances {
		// Skip all-day events from navigation
		if util.IsAllDayEvent(instance.OriginalEvent) {
			continue
		}

		// Check if event matches search criteria
		if !matchesSearchCriteria(instance.OriginalEvent, field, phrase) {
			continue
		}

		eventTime := instance.ComputedStart.Time
		shouldConsider := eventTime.Before(currentTime)
		isBetter := bestInstance == nil || eventTime.After(bestEventTime)

		if shouldConsider && isBetter {
			bestInstance = instance
			bestEventTime = eventTime
		}
	}
	return bestInstance
}
