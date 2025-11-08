package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"

	// popup_forms "github.com/samuelstranges/chronos/popup_forms" // REMOVED
	"github.com/samuelstranges/chronos/popup"
	popup_better_forms "github.com/samuelstranges/chronos/popup_better_forms"
	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// handleSimpleEdit is a helper for simple single-field edits using the new popup_better_forms system
func handleSimpleEdit(m Model, editFunc func(*ical.Event, popup_better_forms.EventUpdateHandler)) (tea.Model, tea.Cmd) {
	// Check if this is a recurring event - block all change operations
	selectedEvent := week_view_grid.GetEventUnderCursor(&m.WeekModel)

	if selectedEvent != nil && selectedEvent.Instance != nil && selectedEvent.Instance.OriginalEvent != nil {
		if allowed, errorMsg := util.ValidateRecurringOperation(selectedEvent.Instance.OriginalEvent, "simple_edit"); !allowed {
			util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, fmt.Errorf(errorMsg), "Action not allowed")
			return m, nil
		}
	}

	if selectedEvent == nil || selectedEvent.Instance == nil || selectedEvent.Instance.OriginalEvent == nil {
		return m, nil
	}

	event := selectedEvent.Instance.OriginalEvent
	editFunc(event, func(properties map[string]string) tea.Cmd {
		return func() tea.Msg {
			err, cmd := m.EventManager.UpdateEvent(event.Component, properties)
			if err != nil {
				return types.VimErrorMsg{Error: fmt.Sprintf("Failed to update event: %v", err)}
			}
			// UpdateEvent now handles refresh automatically via cmd
			if cmd != nil {
				return cmd()
			}
			return types.RefreshMsg{}
		}
	})
	return m, nil
}

// REMOVED: HelperShowChangeForm - replaced by popup_better_forms system

func HandleCalendarPopup(m Model, count int) (tea.Model, tea.Cmd) {
	// Get calendar information and visibility state from EventManager
	calendars := m.EventManager.GetCalendarInfoForForms()
	visibleState := m.EventManager.GetCalendarVisibilityMap()

	// Show the calendar selector popup
	popup.ShowCalendarSelector(calendars, visibleState, func(action popup.CalendarSelectorAction, calendarID string) tea.Cmd {
		return handleCalendarSelectorAction(m, action, calendarID)
	})

	return m, nil
}

// handleCalendarSelectorAction handles actions from the calendar selector popup
func handleCalendarSelectorAction(m Model, action popup.CalendarSelectorAction, calendarID string) tea.Cmd {
	switch action {
	case popup.CalendarActionToggle:
		// Toggle calendar visibility
		currentlyVisible := m.EventManager.IsCalendarVisible(calendarID)
		err, cmd := m.EventManager.SetCalendarVisibility(calendarID, !currentlyVisible)
		if err != nil {
			util.ShowIfError(&m.WeekModel, 5, err, "Failed to toggle calendar visibility")
		}
		return cmd
	case popup.CalendarActionClose:
		// Just close the popup - no action needed
		return nil
	}
	return nil
}

func HandleChangeColor(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditColor)
}

func HandleChangeDuration(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditDuration)
}

func HandleChangeDescription(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditDescription)
}

func HandleChangeTitle(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditTitle)
}

func HandleChangeLocation(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditLocation)
}

func HandleChangeLink(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditLink)
}

func HandleChangeStartTime(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditStartTime)
}

func HandleChangeEndTime(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditEndTime)
}

// HandleEditEvent shows the comprehensive edit event form for non-recurring, timed events
func HandleEditEvent(m Model, count int) (tea.Model, tea.Cmd) {
	return handleSimpleEdit(m, popup_better_forms.ShowEditEvent)
}

// HandleAddEvent shows the add event form
func HandleAddEvent(m Model, count int) (tea.Model, tea.Cmd) {
	// Get calendar information from EventManager
	calendars := m.EventManager.GetCalendarInfoForForms()

	// Show the add event form
	popup_better_forms.ShowAddEvent(&m.WeekModel, calendars, func(properties map[string]string, startTime, endTime time.Time, isAllDay bool, calendarID string) tea.Cmd {
		return func() tea.Msg {
			// If no calendar specified, use the first available calendar
			if calendarID == "" && len(calendars) > 0 {
				calendarID = calendars[0].ID
			}

			// Create the event in the specified calendar using the properties directly
			err, cmd := m.EventManager.CreateEventByCalendarIDWithBatch(calendarID, properties, startTime, endTime, isAllDay)
			if err != nil {
				return types.VimErrorMsg{Error: fmt.Sprintf("Failed to create event: %v", err)}
			}
			// Execute the async sync command (if any)
			if cmd != nil {
				return cmd()
			}
			return types.RefreshMsg{}
		}
	})

	return m, nil
}
