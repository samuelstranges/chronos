package util

import (
	"fmt"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/teambition/rrule-go"
)

// RecurringEventExpander handles the expansion of recurring events into individual occurrences
type RecurringEventExpander struct {
	originalEvent *ical.Component
	rruleSet      *rrule.Set // Use Set instead of RRule to handle EXDATE
	viewStart     time.Time
	viewEnd       time.Time
}

// NewRecurringEventExpander creates a new expander for a recurring event
func NewRecurringEventExpander(eventComponent *ical.Component, viewStart, viewEnd time.Time) (*RecurringEventExpander, error) {
	// Parse RRULE using existing utility
	rule, err := ParseRRuleFromComponent(eventComponent)
	if err != nil {
		return nil, err
	}

	// Get DTSTART using existing utility pattern (convert component to event temporarily)
	event := &ical.Event{Component: eventComponent}
	startLocalTime, err := timezone.GetEventStartTime(event)
	if err != nil {
		return nil, fmt.Errorf("recurring event missing valid DTSTART: %w", err)
	}

	// Convert to time.Time for compatibility with rrule library
	startTime := startLocalTime.Time.In(viewStart.Location())

	// Create a Set to handle both RRULE and EXDATE
	set := &rrule.Set{}
	set.DTStart(startTime)
	set.RRule(rule)

	// Add any existing EXDATE entries
	// Handle both single comma-separated property and multiple EXDATE properties
	exdateProps := eventComponent.Props.Values("EXDATE")
	var allExistingExdates []time.Time
	
	for _, exdateProp := range exdateProps {
		existingExdates, err := ParseExistingExdates(exdateProp.Value, startTime.Location())
		if err != nil {
			return nil, fmt.Errorf("failed to parse EXDATE for recurring event: %w", err)
		}
		allExistingExdates = append(allExistingExdates, existingExdates...)
	}
	
	if len(allExistingExdates) > 0 {
		set.SetExDates(allExistingExdates)
	}

	return &RecurringEventExpander{
		originalEvent: eventComponent,
		rruleSet:      set,
		viewStart:     viewStart,
		viewEnd:       viewEnd,
	}, nil
}

// GenerateOccurrences returns all occurrences within the view timeframe
func (r *RecurringEventExpander) GenerateOccurrences() []*types.EventInstance {
	if r.rruleSet == nil {
		return nil
	}

	// Generate occurrences between viewStart and viewEnd (automatically excludes EXDATE entries)
	occurrences := r.rruleSet.Between(r.viewStart, r.viewEnd, true)

	// Get the original event duration
	duration := GetEventDuration(&ical.Event{Component: r.originalEvent}, r.viewStart.Location())

	var eventInstances []*types.EventInstance
	for _, startTime := range occurrences {
		instance := &types.EventInstance{
			OriginalEvent: &ical.Event{Component: r.originalEvent},
			Calendar:      nil, // Will be set by caller
			ComputedStart: timezone.NewLocalTime(startTime),
			ComputedEnd:   timezone.NewLocalTime(startTime.Add(duration)),
			IsOriginal:    false, // This is a recurring instance
			MasterUID:     GetComponentUID(r.originalEvent),
			InstanceUID:   fmt.Sprintf("%s_%d", GetComponentUID(r.originalEvent), startTime.Unix()),
		}
		eventInstances = append(eventInstances, instance)
	}

	return eventInstances
}

// ExpandRecurringEventIntoInstances expands a single recurring event into event instances within the time range
func ExpandRecurringEventIntoInstances(component *ical.Component, calendar *ical.Calendar, viewStart, viewEnd time.Time) []*types.EventInstance {
	var instances []*types.EventInstance
	
	expander, err := NewRecurringEventExpander(component, viewStart, viewEnd)
	if err != nil {
		return instances // Return empty slice for malformed recurring events
	}

	instances = expander.GenerateOccurrences()
	// Set the calendar reference for all instances
	for _, instance := range instances {
		instance.Calendar = calendar
	}
	
	return instances
}