package ical_crud

import (
	"fmt"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/util"
	"github.com/teambition/rrule-go"
)

// RRuleOperations provides utilities for manipulating RRULE properties
type RRuleOperations struct{}

// NewRRuleOperations creates a new RRULE operations helper
func NewRRuleOperations() *RRuleOperations {
	return &RRuleOperations{}
}

// AddExceptionDate computes EXDATE properties to exclude a specific occurrence from a recurring series
// Returns a property map to be applied via UpdateEvent for proper undo/redo support
func (r *RRuleOperations) AddExceptionDate(component *ical.Component, exceptionDate time.Time) (map[string]string, error) {
	// Get DTSTART to establish the timezone context using existing utility
	event := &ical.Event{Component: component}
	startTime, err := timezone.GetEventStartTime(event)
	if err != nil {
		return nil, fmt.Errorf("recurring event missing valid DTSTART: %w", err)
	}

	// Convert exception date to the same timezone as DTSTART
	exceptionInEventTZ := exceptionDate.In(startTime.Time.Location())

	// Parse existing RRULE using helper
	rule, err := util.ParseRRuleFromComponent(component)
	if err != nil {
		return nil, err
	}

	// Create a Set to handle EXDATE operations
	set := &rrule.Set{}
	set.DTStart(startTime.Time)
	set.RRule(rule)

	// Check if there are existing EXDATEs and add them to the set
	// Handle both single comma-separated property and multiple EXDATE properties
	exdateProps := component.Props.Values("EXDATE")
	var allExistingExdates []time.Time
	
	for _, exdateProp := range exdateProps {
		existingExdates, err := util.ParseExistingExdates(exdateProp.Value, startTime.Time.Location())
		if err != nil {
			return nil, fmt.Errorf("failed to parse existing exception dates: %w", err)
		}
		allExistingExdates = append(allExistingExdates, existingExdates...)
	}
	
	if len(allExistingExdates) > 0 {
		set.SetExDates(allExistingExdates)
	}

	// Add the new exception date
	set.ExDate(exceptionInEventTZ)

	// Compute the new EXDATE property value
	exdateValue := r.updateComponentExdates(component, set.GetExDate())
	
	return map[string]string{
		"EXDATE": exdateValue,
	}, nil
}

// updateComponentExdates computes the EXDATE property value instead of setting it
func (r *RRuleOperations) updateComponentExdates(component *ical.Component, exdates []time.Time) string {
	if len(exdates) == 0 {
		// Return empty string to remove EXDATE property
		return ""
	}

	// Format all exception dates as comma-separated UTC times
	var exdateStrs []string
	for _, exdate := range exdates {
		formatted := exdate.UTC().Format("20060102T150405Z")
		exdateStrs = append(exdateStrs, formatted)
	}

	// Build the EXDATE value string (comma-separated UTC times)
	exdateValue := exdateStrs[0]
	for i := 1; i < len(exdateStrs); i++ {
		exdateValue += "," + exdateStrs[i]
	}

	return exdateValue
}

// TruncateRRuleWithUntil computes RRULE properties to end a recurring series before a specific date
// Returns a property map to be applied via UpdateEvent for proper undo/redo support
func (r *RRuleOperations) TruncateRRuleWithUntil(component *ical.Component, splitDate time.Time) (map[string]string, error) {
	// Parse existing RRULE using helper
	rule, err := util.ParseRRuleFromComponent(component)
	if err != nil {
		return nil, err
	}

	// Get DTSTART to establish the original series start using existing utility
	event := &ical.Event{Component: component}
	startTime, err := timezone.GetEventStartTime(event)
	if err != nil {
		return nil, fmt.Errorf("recurring event missing valid DTSTART: %w", err)
	}

	// Adjust to the split date's timezone context
	startTimeAdjusted := startTime.Time.In(splitDate.Location())

	// Set DTSTART for the rule so we can generate occurrences
	rule.DTStart(startTimeAdjusted)

	// Find the last occurrence before the split date using the rrule library
	lastOccurrenceBeforeSplit := rule.Before(splitDate, false) // false = exclusive

	if lastOccurrenceBeforeSplit.IsZero() {
		// No occurrences before split date - this means we're trying to split before the series starts
		return nil, fmt.Errorf("cannot split recurring series before its start date")
	}

	// Get the original options and set UNTIL to the last valid occurrence
	options := rule.Options
	options.Until = lastOccurrenceBeforeSplit
	options.Count = 0 // COUNT and UNTIL are mutually exclusive per RFC 5545

	// Create new RRULE with updated options
	newRule, err := rrule.NewRRule(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create new RRULE with UNTIL: %w", err)
	}

	// Compute the new RRULE property value
	rruleValue := updateComponentRRule(component, newRule)
	
	return map[string]string{
		"RRULE": rruleValue,
	}, nil
}

// updateComponentRRule computes the RRULE property value instead of setting it
func updateComponentRRule(component *ical.Component, rule *rrule.RRule) string {
	// Use proper ical library method to compute RRULE property value
	rruleString := util.ExtractRRuleStringFromRule(rule)
	return rruleString
}
