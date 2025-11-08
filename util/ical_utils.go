package util

import (
	"fmt"
	"sort"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/types"
)

func IsRecurring(event *ical.Event) bool { return event.Props.Get("RRULE") != nil }

func IsAllDayEvent(event *ical.Event) bool {
	// determines if an event is an all-day event based on iCal specification.
	// According to RFC 5545, all-day events have DTSTART and DTEND properties with VALUE=types.ICalValueTypeDate
	// instead of the default types.ICalValueTypeDate-TIME format.
	if event == nil {
		return false
	}

	dtStartProp := event.Props.Get("DTSTART")
	dtEndProp := event.Props.Get("DTEND")

	// Check DTSTART property for VALUE=types.ICalValueTypeDate parameter
	if dtStartProp == nil {
		return false
	}

	// Check if DTSTART has VALUE=types.ICalValueTypeDate parameter
	valueParam := dtStartProp.Params.Get("VALUE")
	if valueParam == "" || valueParam != types.ICalValueTypeDate {
		return false
	}

	// For consistency, also check DTEND if it exists
	if dtEndProp != nil {
		endValueParam := dtEndProp.Params.Get("VALUE")
		if endValueParam == "" || endValueParam != types.ICalValueTypeDate {
			return false
		}
	}

	return true
}

// getEventProperty returns "" on not found or the property
func getEventPropertyStr(whichProp string, event *ical.Event) string {
	if event == nil || event.Props.Get(whichProp) == nil {
		// property doesn't exist on event or event doesnt exist
		return ""
	}
	return event.Props.Get(whichProp).Value
}
func GetEventUID(event *ical.Event) string         { return getEventPropertyStr("UID", event) }
func GetEventTitle(event *ical.Event) string       { return getEventPropertyStr("SUMMARY", event) }
func GetEventDescription(event *ical.Event) string { return getEventPropertyStr("DESCRIPTION", event) }
func GetEventLocation(event *ical.Event) string    { return getEventPropertyStr("LOCATION", event) }
func GetEventLink(event *ical.Event) string        { return getEventPropertyStr("URL", event) }

// GetComponentUID gets the UID from an iCal component
func GetComponentUID(component *ical.Component) string {
	if component == nil {
		return ""
	}
	if uidProp := component.Props.Get("UID"); uidProp != nil {
		return uidProp.Value
	}
	return ""
}

// ForEachEventInCalendar iterates over all VEVENT components in a calendar
func ForEachEventInCalendar(calendar *ical.Calendar, fn func(*ical.Component) bool) {
	if calendar == nil {
		return
	}
	for _, component := range calendar.Children {
		if component.Name == "VEVENT" {
			if !fn(component) {
				break // Allow early termination
			}
		}
	}
}

// FindEventByUIDInCalendar finds an event by UID within a specific calendar
func FindEventByUIDInCalendar(calendar *ical.Calendar, uid string) (*ical.Component, int) {
	if calendar == nil || uid == "" {
		return nil, -1
	}
	for i, component := range calendar.Children {
		if component.Name == "VEVENT" && GetComponentUID(component) == uid {
			return component, i
		}
	}
	return nil, -1
}

// RemoveEventFromCalendar removes an event by index from calendar.Children
func RemoveEventFromCalendar(calendar *ical.Calendar, index int) bool {
	if calendar == nil || index < 0 || index >= len(calendar.Children) {
		return false
	}
	calendar.Children = append(calendar.Children[:index], calendar.Children[index+1:]...)
	return true
}

// SortEventInstancesChronologically sorts event instances in canonical chronological order
// This ensures every consumer gets consistently sorted events, eliminating picker ordering bugs
func SortEventInstancesChronologically(allInstances []*types.EventInstance) {
	sort.Slice(allInstances, func(i, j int) bool {
		instanceI := allInstances[i]
		instanceJ := allInstances[j]

		// Check if events are all-day by looking at the original event's DTSTART property
		isAllDayI := IsAllDayEvent(instanceI.OriginalEvent)
		isAllDayJ := IsAllDayEvent(instanceJ.OriginalEvent)

		// All-day events come first (conceptually start of day)
		if isAllDayI && !isAllDayJ {
			return true
		}
		if !isAllDayI && isAllDayJ {
			return false
		}

		// Within same type, sort by computed start time
		if !instanceI.ComputedStart.Equal(instanceJ.ComputedStart) {
			return instanceI.ComputedStart.Before(instanceJ.ComputedStart)
		}

		// If same start time, sort by title for deterministic ordering
		iTitle := GetEventTitle(instanceI.OriginalEvent)
		jTitle := GetEventTitle(instanceJ.OriginalEvent)
		return iTitle < jTitle
	})
}

// GetEventColor determines the color for an event with fallback hierarchy:
func GetEventColor(event *ical.Event, calendarColor string) string {
	// 0. Edge case
	if event == nil {
		// Use calendar default or system fallback
		if calendarColor != "" {
			return calendarColor
		}
		return types.DefaultEventColor
	}

	// 1. Check for individual event color first
	if prop := event.Props.Get("X-COLOR"); prop != nil {
		return prop.Value
	}

	// 2. Fall back to calendar default color
	if calendarColor != "" {
		return calendarColor
	}

	// 3. System fallback
	return types.DefaultEventColor
}

// GetCalendarName extracts the display name from a calendar
func GetCalendarName(cal *ical.Calendar) string {
	if prop := cal.Props.Get("X-WR-CALNAME"); prop != nil {
		return prop.Value
	}
	return "Unnamed Calendar"
}

// GetCalendarColor extracts the color from a calendar
func GetCalendarColor(cal *ical.Calendar) string {
	if prop := cal.Props.Get("X-APPLE-CALENDAR-COLOR"); prop != nil {
		return prop.Value
	}
	return types.DefaultEventColor
}

// DeepCopyEvent creates a complete standalone copy of an ical.Event
// This is used for undo functionality and event manipulation
func DeepCopyEvent(original *ical.Event) *ical.Event {
	if original == nil {
		return nil
	}

	copiedComponent := DeepCopyComponent(original.Component)
	return &ical.Event{Component: copiedComponent}
}

// DeepCopyComponent creates a complete standalone copy of an ical.Component
func DeepCopyComponent(original *ical.Component) *ical.Component {
	if original == nil {
		return nil
	}

	// Create new component with same name
	copiedComponent := &ical.Component{
		Name:     original.Name,
		Props:    make(ical.Props),
		Children: make([]*ical.Component, len(original.Children)),
	}

	// Deep copy all properties
	for propName, props := range original.Props {
		copiedComponent.Props[propName] = make([]ical.Prop, len(props))
		for i, prop := range props {
			// Create new property with same values
			newProp := ical.Prop{
				Name:   prop.Name,
				Value:  prop.Value,
				Params: make(ical.Params),
			}
			// Copy parameters
			for paramName, paramValues := range prop.Params {
				newProp.Params[paramName] = append([]string(nil), paramValues...)
			}
			copiedComponent.Props[propName][i] = newProp
		}
	}

	// Deep copy all children recursively
	for i, child := range original.Children {
		copiedComponent.Children[i] = DeepCopyComponent(child)
	}

	return copiedComponent
}

// GetEventDuration calculates the duration of an event using timezone-safe methods
func GetEventDuration(event *ical.Event, defaultLocation *time.Location) time.Duration {
	if event == nil {
		return time.Hour // Default 1 hour
	}

	startTime, endTime, err := timezone.GetEventTimes(event)
	if err != nil {
		return time.Hour // Default 1 hour
	}

	duration := endTime.Sub(startTime)
	if duration <= 0 {
		return time.Hour // Ensure minimum 1 hour duration
	}

	return duration
}

// ValidateRecurringOperation checks if an operation is allowed on a recurring event
// Returns (allowed=true, errorMsg="") for non-recurring events or allowed operations
// Returns (allowed=false, errorMsg="...") for blocked operations on recurring events
func ValidateRecurringOperation(event *ical.Event, operation string) (bool, string) {
	if !IsRecurring(event) {
		return true, "" // Non-recurring events are always allowed
	}

	// Define operations that are NOT allowed on recurring events
	switch operation {
	case "edit", "modify":
		return false, "Modifying recurring events not supported; delete the series and create a new recurring event"
	case "copy":
		return false, "Copying recurring events unsupported"
	case "simple_edit":
		return false, "Modifying recurring events unsupported"
	default:
		return true, "" // Allow delete, view, etc.
	}
}

// NewEventInstanceFromRegular creates an EventInstance from a regular (non-recurring) event
func NewEventInstanceFromRegular(event *ical.Event, calendar *ical.Calendar) (*types.EventInstance, error) {
	startTime, err := timezone.GetEventStartTime(event)
	if err != nil {
		return nil, fmt.Errorf("failed to get event start time: %w", err)
	}

	endTime, err := timezone.GetEventEndTime(event)
	if err != nil {
		return nil, fmt.Errorf("failed to get event end time: %w", err)
	}

	uid := GetEventUID(event)

	return &types.EventInstance{
		OriginalEvent: event,
		Calendar:      calendar,
		ComputedStart: startTime,
		ComputedEnd:   endTime,
		IsOriginal:    true,
		MasterUID:     uid,
		InstanceUID:   uid,
	}, nil
}

// AllDayEventOccursOnDay checks if an all-day event occurs on a specific date.
// Much simpler implementation using existing utilities instead of manual parsing.
func AllDayEventOccursOnDay(event *ical.Event, targetDate time.Time) (bool, error) {
	if !IsAllDayEvent(event) {
		return false, fmt.Errorf("event is not an all-day event")
	}

	startTime, endTime, err := timezone.GetEventTimes(event)
	if err != nil {
		return false, fmt.Errorf("failed to get event times: %w", err)
	}

	// Convert LocalTime to time.Time for compatibility with existing code
	startTimeCompat := startTime.Time
	endTimeCompat := endTime.Time

	// Handle single day event (no end time)
	if endTimeCompat.IsZero() {
		targetDateOnly := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
		startDateOnly := time.Date(startTimeCompat.Year(), startTimeCompat.Month(), startTimeCompat.Day(), 0, 0, 0, 0, time.UTC)
		return targetDateOnly.Equal(startDateOnly), nil
	}

	// Multi-day event - check if target date is within range
	// DTEND is non-inclusive according to RFC 5545
	targetDateOnly := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	startDateOnly := time.Date(startTimeCompat.Year(), startTimeCompat.Month(), startTimeCompat.Day(), 0, 0, 0, 0, time.UTC)
	endDateOnly := time.Date(endTimeCompat.Year(), endTimeCompat.Month(), endTimeCompat.Day(), 0, 0, 0, 0, time.UTC)

	return !targetDateOnly.Before(startDateOnly) && targetDateOnly.Before(endDateOnly), nil
}
