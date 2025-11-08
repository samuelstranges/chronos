package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// HandleCutEvent performs cut operation (copy + delete) on event at cursor
func HandleCutEvent(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager
	if event := week_view_grid.GetEventUnderCursor(weekModel); event != nil {
		// Copy first
		err := eventManager.CopyEvent(event.Instance.OriginalEvent)
		if err == nil {
			// Check if this is a recurring event - if so, return message to show dialog
			if util.IsRecurring(event.Instance.OriginalEvent) {
				// Return a message to trigger the recurring dialog
				return m, func() tea.Msg {
					return types.VimActionMsg{Action: "show_recurring_delete_dialog"}
				}
			}
			// Regular event - direct delete (refresh happens automatically)
			err, cmd := eventManager.DeleteEvent(event.Instance.OriginalEvent)
			util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to delete event")
			if cmd != nil {
				return m, cmd
			}
		}
	}
	return m, nil
}

// HandleCopyEvent performs copy operation on event at cursor
func HandleCopyEvent(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager
	if positionedEvent := week_view_grid.GetEventUnderCursor(weekModel); positionedEvent != nil && positionedEvent.Instance != nil {
		event := positionedEvent.Instance.OriginalEvent
		if event != nil {
			// Block copy operations on recurring events
			if allowed, errorMsg := util.ValidateRecurringOperation(event, "copy"); !allowed {
				// Show error message that copying recurring events is unsupported
				util.ShowIfError(weekModel, types.ErrorDisplaySeconds, fmt.Errorf(errorMsg), "Action not allowed")
				return m, nil
			}
			err := eventManager.CopyEvent(event)
			util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to copy event")
		}
	}
	return m, nil
}

// HandlePasteEvent performs paste operation at cursor position
func HandlePasteEvent(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager
	targetTime := week_view_grid.CellToTimeAtCursor(*weekModel)
	err, cmd := eventManager.PasteEvent(targetTime)
	if err != nil {
		util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to paste event")
		return m, nil
	}
	return m, cmd
}
