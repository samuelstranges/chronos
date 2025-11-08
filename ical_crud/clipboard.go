package ical_crud

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/util"
)

// copyEventToClipboard is a private helper that appends a single event to the clipboard
func (em *EventManager) copyEventToClipboard(event *ical.Event) error {
	// Find which calendar contains this event by matching UID
	eventUID := util.GetEventUID(event)
	if eventUID == "" {
		return fmt.Errorf("event has no UID")
	}

	_, sourceCalendar, _, err := em.FindEventByUID(eventUID)
	if err != nil {
		return fmt.Errorf("could not find source calendar: %w", err)
	}

	// Store reference to original event (no deep copy needed)
	em.copiedEvents = append(em.copiedEvents, CopiedEvent{
		Event:    event,
		Calendar: sourceCalendar,
	})

	return nil
}

// CopyEvents stores multiple events for later pasting (replaces any existing copied events)
func (em *EventManager) CopyEvents(events []*ical.Event) error {
	if len(events) == 0 {
		return fmt.Errorf("no events to copy")
	}

	// Clear existing copied events
	em.copiedEvents = make([]CopiedEvent, 0, len(events))

	// Process each event using the single-event logic
	for _, event := range events {
		if event == nil {
			continue // Skip nil events
		}

		if err := em.copyEventToClipboard(event); err != nil {
			// Continue with other events, but track if we have any valid ones
			continue
		}
	}

	if len(em.copiedEvents) == 0 {
		return fmt.Errorf("no valid events found to copy")
	}

	return nil
}

// CopyEvent stores a single event for later pasting (replaces any existing copied events)
func (em *EventManager) CopyEvent(event *ical.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	// Clear existing copied events for single event copy
	em.copiedEvents = make([]CopiedEvent, 0, 1)
	return em.copyEventToClipboard(event)
}

// HasRecurringEventInClipboard checks if any copied events are recurring events
func (em *EventManager) HasRecurringEventInClipboard() bool {
	for _, copiedEvent := range em.copiedEvents {
		if util.IsRecurring(copiedEvent.Event) {
			return true
		}
	}
	return false
}

// PasteEvent creates new events at the specified position using the copied events data
// Maintains relative positioning between events from when they were copied
func (em *EventManager) PasteEvent(targetTime time.Time) (error, tea.Cmd) {
	if len(em.copiedEvents) == 0 {
		return fmt.Errorf("no events copied"), nil
	}

	// Block paste operations if clipboard contains recurring events
	if em.HasRecurringEventInClipboard() {
		return fmt.Errorf("cannot paste recurring events - copy/paste not supported for recurring events"), nil
	}

	// Use the provided target time for positioning
	baseTime := targetTime

	// Find the earliest start time among all copied events to use as reference
	var earliestStart time.Time
	for i, copiedEvent := range em.copiedEvents {
		eventStart, _, err := timezone.GetEventTimes(copiedEvent.Event)
		if err != nil {
			continue // Skip events with invalid times
		}
		if !eventStart.IsZero() {
			if i == 0 || eventStart.Time.Before(earliestStart) {
				earliestStart = eventStart.Time
			}
		}
	}

	// Collect all paste changes for batch operation
	var changes []SingleEventChange
	description := fmt.Sprintf("Paste %d event(s)", len(em.copiedEvents))

	// Paste each copied event maintaining relative position from earliest start
	for _, copiedEvent := range em.copiedEvents {
		newEvent, err := em.createPastedEvent(copiedEvent, baseTime, earliestStart)
		if err != nil {
			return fmt.Errorf("failed to paste event: %w", err), nil
		}

		// Add to batch changes instead of recording individually
		changes = append(changes, SingleEventChange{
			EventData: newEvent,
			Calendar:  copiedEvent.Calendar,
		})
	}

	// Store the change data
	eventsChanged := EventsChanged{Changes: changes}

	// Record all paste operations as a single batch and return cmd (handles async sync + refresh)
	return nil, em.recordChange(util.ChangeTypeAdd, eventsChanged, description)
}

// createPastedEvent creates a pasted event maintaining its relative position from the earliest event
func (em *EventManager) createPastedEvent(copiedEvent CopiedEvent, baseTime time.Time, earliestStart time.Time) (*ical.Event, error) {
	// Deep copy the event and modify what we need
	newEvent := util.DeepCopyEvent(copiedEvent.Event)

	// Get original times and check if all-day
	originalStart, originalEnd, err := timezone.GetEventTimes(copiedEvent.Event)
	if err != nil {
		return nil, fmt.Errorf("failed to get event times: %w", err)
	}
	isAllDay := util.IsAllDayEvent(copiedEvent.Event)

	// Calculate new times with offset
	duration := originalEnd.Time.Sub(originalStart.Time)
	offsetFromEarliest := originalStart.Time.Sub(earliestStart)
	newStartTime := baseTime.In(originalStart.Time.Location()).Add(offsetFromEarliest)
	newEndTime := newStartTime.Add(duration)

	// Generate new UID for the copied event
	newUID, err := util.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new UID: %w", err)
	}
	newEvent.Props.SetText("UID", newUID)

	// Set new times using timezone-safe methods to preserve original timezone format
	if isAllDay {
		newEvent.Props.SetDate("DTSTART", newStartTime)
		newEvent.Props.SetDate("DTEND", newEndTime)
	} else {
		// Get original format from source event to preserve timezone format
		originalStartProp := copiedEvent.Event.Props.Get("DTSTART")
		originalEndProp := copiedEvent.Event.Props.Get("DTEND")
		
		originalStartValue := ""
		if originalStartProp != nil {
			originalStartValue = originalStartProp.Value
		}
		originalEndValue := ""
		if originalEndProp != nil {
			originalEndValue = originalEndProp.Value
		}
		
		// Use timezone-safe methods to preserve original timezone format
		timezone.SetEventDateTimeProperty(newEvent, "DTSTART", 
			timezone.NewLocalTime(newStartTime), originalStartValue)
		timezone.SetEventDateTimeProperty(newEvent, "DTEND", 
			timezone.NewLocalTime(newEndTime), originalEndValue)
	}

	// Add directly to target calendar
	copiedEvent.Calendar.Children = append(copiedEvent.Calendar.Children, newEvent.Component)

	// Return the pasted event for batch recording
	return newEvent, nil
}
