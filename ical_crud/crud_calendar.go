// Package eventmanager handles calendar and event operations including CRUD operations and data management.
package ical_crud

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// Calendar-related operations extracted from eventmanager.go

// buildCalendarMap - REMOVED! 
// Now using stable IDs directly from storage via rebuildCalendarSystem()

// GetCalendarInfoForForms returns calendar info (ID, name, and color) for form selectors
func (em *EventManager) GetCalendarInfoForForms() []types.CalendarInfo {
	var infos []types.CalendarInfo
	for id, cal := range em.calendarMap {
		name := "Unnamed Calendar"
		if prop := cal.Props.Get("X-WR-CALNAME"); prop != nil {
			name = prop.Value
		}
		color := util.GetCalendarColor(cal)
		infos = append(infos, types.CalendarInfo{ID: id, Name: name, Color: color})
	}
	return infos
}

// GetCalendarsForDisplay returns only visible calendars for display purposes
func (em *EventManager) GetCalendarsForDisplay() []*ical.Calendar {
	var visibleCalendars []*ical.Calendar
	for calendarID, calendar := range em.calendarMap {
		if em.IsCalendarVisible(calendarID) {
			visibleCalendars = append(visibleCalendars, calendar)
		}
	}
	return visibleCalendars
}

// IsCalendarVisible checks if a calendar is visible
func (em *EventManager) IsCalendarVisible(calendarID string) bool {
	visible, exists := em.calendarVisible[calendarID]
	return exists && visible // Default to false if not found
}

// GetCalendarVisibilityMap returns a map of calendar IDs to their visibility state
func (em *EventManager) GetCalendarVisibilityMap() map[string]bool {
	visibilityMap := make(map[string]bool)
	for calendarID := range em.calendarMap {
		visibilityMap[calendarID] = em.IsCalendarVisible(calendarID)
	}
	return visibilityMap
}

// SetCalendarVisibility sets calendar visibility and returns refresh command
// Note: Visibility is local UI state only, not persisted to storage
func (em *EventManager) SetCalendarVisibility(calendarID string, visible bool) (error, tea.Cmd) {
	if _, exists := em.calendarMap[calendarID]; exists {
		em.calendarVisible[calendarID] = visible
		// No storage sync needed - visibility is local UI state only
		return nil, func() tea.Msg { return types.RefreshMsg{} }
	}
	return fmt.Errorf("calendar %s not found", calendarID), nil
}

// GetCalendarNameForEvent finds the calendar name for a given event
func (em *EventManager) GetCalendarNameForEvent(event *ical.Event) string {
	if event == nil {
		return types.UnknownCalendar
	}

	// Get the event's UID to find its source calendar
	eventUID := ""
	if uidProp := event.Props.Get("UID"); uidProp != nil {
		eventUID = uidProp.Value
	} else {
		return types.UnknownCalendar
	}

	// Search through all calendars to find the one containing this event
	for _, calendar := range em.calendars {
		if calendar == nil {
			continue
		}

		component, _ := util.FindEventByUIDInCalendar(calendar, eventUID)
		if component != nil {
			return util.GetCalendarName(calendar)
		}
	}

	return "Unknown Calendar"
}

