package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/popup"
	popup_better_forms "github.com/samuelstranges/chronos/popup_better_forms"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// openEventInstanceURL opens URL for a specific EventInstance
func openEventInstanceURL(m Model, instance *types.EventInstance) (tea.Model, tea.Cmd) {
	err := openICalEventURL(instance.OriginalEvent)
	util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Cant show url")
	return m, nil
}

// handleAgendaPickerAction handles actions from the agenda picker popup using EventInstance methods
func handleAgendaPickerAction(m Model, action popup.AgendaPickerAction, event *ical.Event, dayEvents []*types.EventInstance) tea.Cmd {
	// Find the EventInstance that matches the selected event
	var selectedInstance *types.EventInstance
	for _, instance := range dayEvents {
		if instance.OriginalEvent == event {
			selectedInstance = instance
			break
		}
	}

	if selectedInstance == nil {
		return nil
	}

	switch action {
	case popup.AgendaActionViewDetails:
		return func() tea.Msg {
			// Call the function directly in the message - this executes in main update loop
			calendarName := m.EventManager.GetCalendarNameForEvent(selectedInstance.OriginalEvent)
			popup.ShowEventDetailsPopup(selectedInstance.OriginalEvent, calendarName)
			return types.RefreshMsg{}
		}
	case popup.AgendaActionEdit:
		return func() tea.Msg {
			// Call the function directly in the message - this executes in main update loop
			if allowed, errorMsg := util.ValidateRecurringOperation(selectedInstance.OriginalEvent, "simple_edit"); !allowed {
				return types.VimErrorMsg{Error: errorMsg}
			}


			popup_better_forms.ShowEditEvent(selectedInstance.OriginalEvent, func(properties map[string]string) tea.Cmd {
				return func() tea.Msg {
					err, cmd := m.EventManager.UpdateEvent(selectedInstance.OriginalEvent.Component, properties)
					if err != nil {
						return types.VimErrorMsg{Error: fmt.Sprintf("Failed to update event: %v", err)}
					}
					if cmd != nil {
						return cmd()
					}
					return types.RefreshMsg{}
				}
			})
			return types.RefreshMsg{}
		}
	case popup.AgendaActionDelete:
		// Handle delete - check for recurring events
		if util.IsRecurring(event) {
			return func() tea.Msg {
				return types.VimActionMsg{Action: "show_recurring_delete_dialog_from_picker"}
			}
		} else {
			// Direct delete for non-recurring events
			err, cmd := m.EventManager.DeleteEvent(event)
			if err != nil {
				util.ShowIfError(&m.WeekModel, 5, err, "Delete failed")
			}
			return cmd
		}
	case popup.AgendaActionOpenURL:
		_, cmd := openEventInstanceURL(m, selectedInstance)
		return cmd
	case popup.AgendaActionClose:
		// Just close the popup - no action needed
		return nil
	}
	return nil
}
