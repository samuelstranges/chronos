package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/ical_crud"
	"github.com/samuelstranges/chronos/popup"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// HandleShowRecuringDeleteDialog shows recurring delete dialog for event under cursor
func HandleShowRecuringDeleteDialog(m Model, count int) (tea.Model, tea.Cmd) {
	selectedEvent := week_view_grid.GetEventUnderCursor(&m.WeekModel)
	if selectedEvent != nil && selectedEvent.Instance != nil && selectedEvent.Instance.OriginalEvent != nil {
		showRecurringDialog(selectedEvent.Instance.OriginalEvent, types.OperationTypeDelete)
	}
	return m, nil
}

// HandleRecurringDeleteFromPicker shows recurring delete dialog for selected event from picker
func HandleRecurringDeleteFromPicker(m Model, count int) (tea.Model, tea.Cmd) {
	if m.WeekModel.SelectedEvent != nil {
		showRecurringDialog(m.WeekModel.SelectedEvent, types.OperationTypeDelete)
		m.WeekModel.SelectedEvent = nil
	} else if m.WeekModel.SelectedAllDayEvent != nil {
		showRecurringDialog(m.WeekModel.SelectedAllDayEvent, types.OperationTypeDelete)
		m.WeekModel.SelectedAllDayEvent = nil
	}
	return m, nil
}

// showRecurringDialog displays the recurring dialog popup
func showRecurringDialog(event *ical.Event, operationType types.RecurringEventOperationType) {
	popup.ShowRecurringDialog(event, operationType, func(choice popup.RecurringDialogChoice) tea.Cmd {
		return func() tea.Msg {
			var scope types.RecurringEventOperationScope
			cancelled := false

			switch choice {
			case popup.ChoiceThis:
				scope = types.OperationScopeThisOnly
			case popup.ChoiceFuture:
				scope = types.OperationScopeThisAndFuture
			case popup.ChoiceAll:
				scope = types.OperationScopeAll
			case popup.ChoiceCancel:
				cancelled = true
			}

			return types.RecurringDialogChoiceMsg{
				Event:         event,
				Scope:         scope,
				OperationType: operationType,
				Cancelled:     cancelled,
			}
		}
	})
}

// handleRecurringDialogChoice processes the user's choice from the recurring dialog
func handleRecurringDialogChoice(m Model, msg types.RecurringDialogChoiceMsg) (tea.Model, tea.Cmd) {
	// Handle cancellation
	if msg.Cancelled {
		return m, nil
	}

	// Only handle delete operations for now (edit operations would be different)
	if msg.OperationType != types.OperationTypeDelete {
		util.ShowIfError(&m.WeekModel, 5, fmt.Errorf("unsupported operation type"), "Operation not supported")
		return m, nil
	}

	// Handle deletion based on scope
	switch msg.Scope {
	case types.OperationScopeThisOnly:
		return handleDeleteThisOnly(m, msg.Event)
	case types.OperationScopeThisAndFuture:
		return handleDeleteThisAndFuture(m, msg.Event)
	case types.OperationScopeAll:
		return handleDeleteAll(m, msg.Event)
	default:
		util.ShowIfError(&m.WeekModel, 5, fmt.Errorf("unknown scope"), "Unknown operation scope")
		return m, nil
	}
}

// handleDeleteThisOnly adds an EXDATE to exclude this specific occurrence
func handleDeleteThisOnly(m Model, event *ical.Event) (tea.Model, tea.Cmd) {
	// Get the current occurrence date from cursor position
	selectedEventInfo := week_view_grid.GetEventUnderCursor(&m.WeekModel)
	if selectedEventInfo == nil || selectedEventInfo.Instance == nil {
		util.ShowIfError(&m.WeekModel, 5, fmt.Errorf("no event selected"), "No event found")
		return m, nil
	}

	// Use the computed start time of the selected instance
	occurrenceDate := selectedEventInfo.Instance.ComputedStart.Time

	// Compute the new EXDATE property value
	rruleOps := ical_crud.NewRRuleOperations()
	properties, err := rruleOps.AddExceptionDate(event.Component, occurrenceDate)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to compute exception date")
		return m, nil
	}

	// Apply the computed properties via UpdateEvent (proper undo/redo support)
	err, cmd := m.EventManager.UpdateEvent(event.Component, properties)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to save changes")
		return m, nil
	}

	return m, cmd
}

// handleDeleteThisAndFuture modifies the RRULE to end before this occurrence
func handleDeleteThisAndFuture(m Model, event *ical.Event) (tea.Model, tea.Cmd) {
	// Get the current occurrence date from cursor position
	selectedEventInfo := week_view_grid.GetEventUnderCursor(&m.WeekModel)
	if selectedEventInfo == nil || selectedEventInfo.Instance == nil {
		util.ShowIfError(&m.WeekModel, 5, fmt.Errorf("no event selected"), "No event found")
		return m, nil
	}

	// Use the computed start time of the selected instance
	splitDate := selectedEventInfo.Instance.ComputedStart.Time

	// Compute the new RRULE property value
	rruleOps := ical_crud.NewRRuleOperations()
	properties, err := rruleOps.TruncateRRuleWithUntil(event.Component, splitDate)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to compute truncated recurring series")
		return m, nil
	}

	// Apply the computed properties via UpdateEvent (proper undo/redo support)
	err, cmd := m.EventManager.UpdateEvent(event.Component, properties)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to save changes")
		return m, nil
	}

	return m, cmd
}

// handleDeleteAll deletes the entire recurring series
func handleDeleteAll(m Model, event *ical.Event) (tea.Model, tea.Cmd) {
	// Simply delete the event entirely
	err, cmd := m.EventManager.DeleteEvent(event)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to delete event")
		return m, nil
	}

	return m, cmd
}

