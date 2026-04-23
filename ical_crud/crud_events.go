package ical_crud

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// getEventsForDateRange returns a unified collection of EventInstances for the given date range
// This includes both regular events and expanded recurring event instances from visible calendars only
func (em *EventManager) GetEventsForDateRange(calendarIDs []string, viewStart, viewEnd time.Time) ([]*types.EventInstance, error) {
	var allInstances []*types.EventInstance

	// Use only visible calendars instead of the provided calendarIDs
	visibleCalendars := em.GetCalendarsForDisplay()
	
	for _, calendar := range visibleCalendars {
		util.ForEachEventInCalendar(calendar, func(component *ical.Component) bool {
			if util.IsRecurring(&ical.Event{Component: component}) {
				instances := util.ExpandRecurringEventIntoInstances(component, calendar, viewStart, viewEnd)
				allInstances = append(allInstances, instances...)
			} else {
				// Regular event
				event := &ical.Event{Component: component}
				instance, err := util.NewEventInstanceFromRegular(event, calendar)
				if err != nil {
					return true // Skip malformed events
				}

				// Only include if event is within the week
				if util.EventInstanceOverlapsTimeRange(instance, viewStart, viewEnd) {
					allInstances = append(allInstances, instance)
				}
			}
			return true // Continue processing
		})
	}

	// Sort events chronologically using canonical sort
	util.SortEventInstancesChronologically(allInstances)
	return allInstances, nil
}

// GetAllEventsForWeek returns all events (both all-day and timed) for the specified week from visible calendars
func (em *EventManager) GetAllEventsForWeek(weekModel *types.WeekModel) ([]*types.EventInstance, error) {
	// Calculate week start and end times
	weekStart := weekModel.CurrentlyViewedWeek
	weekEnd := weekStart.AddDate(0, 0, types.DaysPerWeek)
	
	// Use GetEventsForDateRange (calendarIDs parameter is ignored as we use visible calendars internally)
	return em.GetEventsForDateRange(nil, weekStart, weekEnd)
}

// GetTimedEventsForWeek returns only timed (non-all-day) events for the specified week from visible calendars
func (em *EventManager) GetTimedEventsForWeek(weekModel *types.WeekModel) ([]*types.EventInstance, error) {
	allEvents, err := em.GetAllEventsForWeek(weekModel)
	if err != nil {
		return nil, err
	}
	
	var timedEvents []*types.EventInstance
	for _, instance := range allEvents {
		if !util.IsAllDayEvent(instance.OriginalEvent) {
			timedEvents = append(timedEvents, instance)
		}
	}
	
	return timedEvents, nil
}

// GetAllDayEventsForWeek returns only all-day events for the specified week from visible calendars
func (em *EventManager) GetAllDayEventsForWeek(weekModel *types.WeekModel) ([]*types.EventInstance, error) {
	allEvents, err := em.GetAllEventsForWeek(weekModel)
	if err != nil {
		return nil, err
	}
	
	var allDayEvents []*types.EventInstance
	for _, instance := range allEvents {
		if util.IsAllDayEvent(instance.OriginalEvent) {
			allDayEvents = append(allDayEvents, instance)
		}
	}
	
	return allDayEvents, nil
}

// FindEventByUID searches through EventManager's calendars to find an event with the given UID
// Returns the event, its containing calendar, component index, and any error
func (em *EventManager) FindEventByUID(uid string) (*ical.Event, *ical.Calendar, int, error) {
	if uid == "" {
		return nil, nil, -1, fmt.Errorf("UID cannot be empty")
	}

	for _, calendar := range em.calendars {
		if calendar == nil {
			continue
		}
		component, i := util.FindEventByUIDInCalendar(calendar, uid)
		if component != nil {
			return &ical.Event{Component: component}, calendar, i, nil
		}
	}

	return nil, nil, -1, fmt.Errorf("event with UID '%s' not found", uid)
}

// Event-related CRUD operations extracted from eventmanager.go

// CreateEventByCalendarIDWithBatch creates a new event with arbitrary properties and batch control
// Returns error and a tea.Cmd for async server sync
func (em *EventManager) CreateEventByCalendarIDWithBatch(
	calendarID string,
	properties map[string]string,
	startDateTime, endDateTime time.Time,
	isAllDay bool,
) (error, tea.Cmd) {
	calendar, exists := em.calendarMap[calendarID]
	if !exists {
		return fmt.Errorf("calendar with ID '%s' not found", calendarID), nil
	}

	cmd, err := em.createEventInCalendarWithProperties(
		calendar, properties, startDateTime, endDateTime, isAllDay,
	)
	if err != nil {
		return err, nil
	}

	// Return the cmd from recordChange (handles async sync + refresh)
	return nil, cmd
}

// createEventInCalendarWithProperties creates a new event with arbitrary properties and adds it to the specified calendar
// Returns tea.Cmd for async server sync
func (em *EventManager) createEventInCalendarWithProperties(
	calendar *ical.Calendar,
	properties map[string]string,
	startDateTime, endDateTime time.Time,
	isAllDay bool,
) (tea.Cmd, error) {
	if calendar == nil {
		return nil, fmt.Errorf("calendar cannot be nil")
	}

	// Create new iCal event
	event := ical.NewEvent()

	// Ensure event has required UID
	if err := ensureEventUID(event); err != nil {
		return nil, err
	}

	// Set DTSTAMP (required by RFC 5545)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())

	// Set time properties based on all-day flag
	setEventTimeProperties(event, startDateTime, endDateTime, isAllDay)

	// Set custom properties from map
	setEventCustomProperties(event, properties)

	// Add event to calendar
	calendar.Children = append(calendar.Children, event.Component)

	// Store the change data
	eventsChanged := EventsChanged{
		Changes: []SingleEventChange{{
			EventData: event,
			Calendar:  calendar,
		}},
	}

	// Record the change and return cmd (handles async sync + refresh)
	return em.recordChange(util.ChangeTypeAdd, eventsChanged, "Create event"), nil
}

// replaceEventComponentByUID replaces an event component by UID (core operation without side effects)
// Returns the old component and true if successful, nil and false if event not found
func (em *EventManager) replaceEventComponentByUID(uid string, newComponent *ical.Component) (*ical.Component, bool) {
	if uid == "" || newComponent == nil {
		return nil, false
	}

	_, calendar, index, err := em.FindEventByUID(uid)
	if err != nil {
		return nil, false
	}

	// Store old component for return
	oldComponent := calendar.Children[index]

	// Replace with new component
	calendar.Children[index] = newComponent

	return oldComponent, true
}

// deleteEventByUID removes a single event by UID and records the change (without committing)
// Returns true if event was found and deleted, false otherwise
func (em *EventManager) deleteEventByUID(uid string) (*ical.Event, bool) {
	if uid == "" {
		return nil, false
	}

	// Use existing FindEventByUID method
	event, calendar, index, err := em.FindEventByUID(uid)
	if err != nil {
		return nil, false
	}

	// Remove the event from the calendar
	util.RemoveEventFromCalendar(calendar, index)
	return event, true
}

// DeleteEvent removes an event from its calendar and records the change
func (em *EventManager) DeleteEvent(event *ical.Event) (error, tea.Cmd) {
	// Get the UID of the event we want to delete
	targetUID := util.GetEventUID(event)
	if targetUID == "" {
		return fmt.Errorf("event has no UID"), nil
	}

	// Find the event and calendar before deleting
	_, calendar, _, err := em.FindEventByUID(targetUID)
	if err != nil {
		return fmt.Errorf("event not found"), nil
	}

	// Delete the event
	_, deleted := em.deleteEventByUID(targetUID)
	if !deleted {
		return fmt.Errorf("event not found"), nil
	}

	// Store the change data
	eventsChanged := EventsChanged{
		Changes: []SingleEventChange{{
			EventData: event,
			Calendar:  calendar,
		}},
	}

	// Record the change and return cmd (handles async sync + refresh)
	return nil, em.recordChange(util.ChangeTypeDelete, eventsChanged, "Delete event")
}

// DeleteEvents removes multiple events and records them as a single batch operation
func (em *EventManager) DeleteEvents(events []*ical.Event) (error, tea.Cmd) {
	if len(events) == 0 {
		return fmt.Errorf("no events to delete"), nil
	}

	// Collect all delete changes for batch operation
	var changes []SingleEventChange
	description := fmt.Sprintf("Delete %d event(s)", len(events))

	// Delete each event and collect changes
	for _, event := range events {
		// Get the UID of the event we want to delete
		targetUID := util.GetEventUID(event)
		if targetUID == "" {
			continue // Skip events without UID
		}

		// Find the calendar before deleting
		_, calendar, _, err := em.FindEventByUID(targetUID)
		if err != nil {
			continue // Skip events not found
		}

		// Delete the event
		_, deleted := em.deleteEventByUID(targetUID)
		if deleted {
			// Add to batch changes
			changes = append(changes, SingleEventChange{
				EventData: event,
				Calendar:  calendar,
			})
		}
	}

	if len(changes) == 0 {
		return fmt.Errorf("no events were deleted"), nil
	}

	// Store the change data
	eventsChanged := EventsChanged{Changes: changes}

	// Record all deletions as a single batch and return cmd (handles async sync + refresh)
	return nil, em.recordChange(util.ChangeTypeDelete, eventsChanged, description)
}

// UpdateEvent modifies an existing event's properties and records the change
func (em *EventManager) UpdateEvent(eventComponent *ical.Component, properties map[string]string) (error, tea.Cmd) {
	if eventComponent == nil {
		return fmt.Errorf("event component cannot be nil"), nil
	}
	

	// Validate required properties
	targetUID := util.GetComponentUID(eventComponent)
	if !util.IsValidUID(targetUID) {
		return fmt.Errorf("event has no UID"), nil
	}
	if dtstart, exists := properties["DTSTART"]; exists && !timezone.IsValidStartTime(dtstart) {
		return fmt.Errorf("invalid DTSTART format: %s", dtstart), nil
	}
	if dtend, exists := properties["DTEND"]; exists && !timezone.IsValidEndTime(dtend) {
		return fmt.Errorf("invalid DTEND format: %s", dtend), nil
	}
	if err := timezone.ValidateICalTimeRange(properties["DTSTART"], properties["DTEND"]); err != nil {
		return fmt.Errorf("invalid time range: %w", err), nil
	}

	// Store DEEP COPY of the original event for undo functionality
	originalEvent := util.DeepCopyEvent(&ical.Event{Component: eventComponent})

	// Update all properties from the map (includes DTSTART, DTEND, etc.)
	setEventCustomProperties(&ical.Event{Component: eventComponent}, properties)

	// Use shared method to replace the component
	_, replaced := em.replaceEventComponentByUID(targetUID, eventComponent)
	if !replaced {
		return fmt.Errorf("could not find event to update"), nil
	}

	// Get the calendar for recording the change
	_, sourceCalendar, _, _ := em.FindEventByUID(targetUID)

	// Create modified event state for redo (after all changes are applied)
	modifiedEvent := util.DeepCopyEvent(&ical.Event{Component: eventComponent})


	// Store the change data
	eventsChanged := EventsChanged{
		Changes: []SingleEventChange{{
			EventData:         modifiedEvent,
			PreviousEventData: originalEvent,
			Calendar:          sourceCalendar,
		}},
	}

	// Record the change and return cmd (handles async sync + refresh)
	return nil, em.recordChange(util.ChangeTypeEdit, eventsChanged, "Edit event")
}

// setEventTimeProperties sets DTSTART and DTEND properties using timezone-safe methods
func setEventTimeProperties(event *ical.Event, startDateTime, endDateTime time.Time, isAllDay bool) {
	if isAllDay {
		// All-day events use DATE format (no time, VALUE=DATE)
		event.Props.SetDate("DTSTART", startDateTime)
		event.Props.SetDate("DTEND", endDateTime)
	} else {
		// Timed events - use timezone-safe methods to prevent timezone corruption
		startLocalTime := timezone.NewLocalTime(startDateTime)
		endLocalTime := timezone.NewLocalTime(endDateTime)
		
		// Set properties using timezone-safe methods that preserve local time format
		timezone.SetEventDateTimeProperty(event, "DTSTART", startLocalTime, "")
		timezone.SetEventDateTimeProperty(event, "DTEND", endLocalTime, "")
	}
}

// setEventCustomProperties sets custom properties from a map, skipping empty values
// Handles datetime properties correctly with SetDateTime/SetDate vs SetText
func setEventCustomProperties(event *ical.Event, properties map[string]string) {
	for propName, propValue := range properties {
		if propValue == "" {
			// Remove property if empty
			event.Props.Del(propName)
			continue
		}

		// Handle different property types
		switch propName {
		case "DTSTART", "DTEND", "CREATED", "LAST-MODIFIED", "DTSTAMP":
			// DateTime properties - use timezone-safe validation and setting
			if timezone.IsValidICalDateTime(propValue) || timezone.IsValidICalDate(propValue) {
				// CRITICAL: All-day event property handling
				// 
				// All-day events MUST have VALUE=DATE parameter to be recognized:
				// - DTSTART:20241221 (without VALUE=DATE) = broken, appears as timed event
				// - DTSTART;VALUE=DATE:20241221 = correct all-day event
				//
				// Bug was: SetText() overwrote properties without preserving VALUE=DATE
				// Fix: Detect DATE-only values and explicitly set VALUE=DATE parameter
				if timezone.IsValidICalDate(propValue) && len(propValue) == 8 {
					// This is a DATE value (YYYYMMDD format) - set with VALUE=DATE parameter
					event.Props.Set(&ical.Prop{
						Name:  propName,
						Value: propValue,
						Params: ical.Params{
							"VALUE": []string{"DATE"}, // This makes it an all-day event!
						},
					})
				} else {
					// Regular datetime - set directly so we don't get VALUE=TEXT
					// (SetText escapes semicolons and emits VALUE=TEXT, which
					// breaks DTSTART/DTEND/DTSTAMP/LAST-MODIFIED parsing).
					event.Props.Set(&ical.Prop{
						Name:  propName,
						Value: propValue,
					})
				}
			} else {
				// If we can't parse as datetime, fall back to text (for backward compatibility)
				event.Props.SetText(propName, propValue)
			}
		case "EXDATE":
			// EXDATE is a structured property like RRULE - must not use SetText
			// SetText adds VALUE=TEXT which is invalid
			// propValue can be comma-separated UTC times (e.g., "20060102T150405Z,20070102T150405Z")
			event.Props.Set(&ical.Prop{
				Name:  propName,
				Value: propValue,
			})
		case "RRULE":
			// RRULE is a structured property - must not use SetText
			// SetText adds VALUE=TEXT which is invalid for RRULE and causes CalDAV rejection
			event.Props.Set(&ical.Prop{
				Name:  propName,
				Value: propValue,
			})
		default:
			// Text properties (SUMMARY, DESCRIPTION, LOCATION, URL, X-COLOR, etc.)
			event.Props.SetText(propName, propValue)
		}
	}
}

// ensureEventUID ensures the event has a UID, generating one if missing
func ensureEventUID(event *ical.Event) error {
	if util.GetEventUID(event) == "" {
		uid, err := util.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate event UID: %w", err)
		}
		event.Props.SetText("UID", uid)
	}
	return nil
}

// GetEventsForDayUnderCursor returns all events for the day under cursor
func (em *EventManager) GetEventsForDayUnderCursor(weekModel *types.WeekModel) ([]*types.EventInstance, error) {
	dayIndex := weekModel.Cursor.Day
	dayStart := weekModel.CurrentlyViewedWeek.AddDate(0, 0, dayIndex)
	dayEnd := dayStart.AddDate(0, 0, 1)
	
	// Query a wider range to catch potential multi-day events
	queryStart := dayStart.AddDate(0, 0, -1) // Start from previous day
	queryEnd := dayStart.AddDate(0, 0, 2)    // End after next day
	
	allEvents, err := em.GetEventsForDateRange(nil, queryStart, queryEnd)
	if err != nil {
		return nil, err
	}
	
	// Filter using the same overlap logic as grid (RFC 5545 compliant)
	// This naturally handles boundary cases and multi-day events correctly
	var dayEvents []*types.EventInstance
	for _, instance := range allEvents {
		if util.EventInstanceOverlapsTimeRange(instance, dayStart, dayEnd) {
			dayEvents = append(dayEvents, instance)
		}
	}
	
	return dayEvents, nil
}

// GetAllDayEventsForDayUnderCursor returns all-day events for the day under cursor
func (em *EventManager) GetAllDayEventsForDayUnderCursor(weekModel *types.WeekModel) ([]*types.EventInstance, error) {
	dayEvents, err := em.GetEventsForDayUnderCursor(weekModel)
	if err != nil {
		return nil, err
	}
	
	var allDayEvents []*types.EventInstance
	for _, instance := range dayEvents {
		if util.IsAllDayEvent(instance.OriginalEvent) {
			allDayEvents = append(allDayEvents, instance)
		}
	}
	
	return allDayEvents, nil
}
