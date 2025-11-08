package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// HandleEnterVisualMode enters visual mode and sets anchor
func HandleEnterVisualMode(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel // Use pointer to modify original
	vimState := m.VimState

	vimState.SetMode(types.ModeVisual)
	vimState.SetVisualAnchor(weekModel.Cursor)
	weekModel.IsVisualMode = true
	weekModel.VisualAnchor = weekModel.Cursor
	weekModel.VisualAnchorWeek = weekModel.CurrentlyViewedWeek // Track which week anchor belongs to
	return m, nil
}

// HandleExitVisualMode exits visual mode back to normal
func HandleExitVisualMode(m Model, count int) (tea.Model, tea.Cmd) {
	m.VimState.SetMode(types.ModeNormal)
	m.WeekModel.IsVisualMode = false
	return m, nil
}

// HandleCopySelection copies all events in visual selection
func HandleCopySelection(m Model, count int) (tea.Model, tea.Cmd) {
	// IMPORTANT: Get events BEFORE changing mode - visual selection function checks IsVisualMode
	selectedEvents := week_view_grid.GetEventsInVisualSelectionByTime(&m.WeekModel, m.EventManager)

	// Now exit visual mode
	m.VimState.SetMode(types.ModeNormal)
	m.WeekModel.IsVisualMode = false
	if len(selectedEvents) > 0 {
		// Convert SafeEvents to ical.Components and check for recurring events
		var events []*ical.Event
		var hasRecurringEvents bool
		for _, event := range selectedEvents {
			// Check if any selected event is recurring
			if util.IsRecurring(event) {
				hasRecurringEvents = true
				continue // Skip recurring events
			}
			events = append(events, event)
		}

		// Show error if any recurring events were in selection
		if hasRecurringEvents {
			util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, fmt.Errorf("unsupported operation"), "Copying recurring events")
		}

		// Copy only non-recurring events
		if len(events) > 0 {
			err := m.EventManager.CopyEvents(events)
			util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Failed to copy events")
		}
	}
	m.VimState.SetMode(types.ModeNormal)
	m.WeekModel.IsVisualMode = false
	return m, nil
}

// HandleCutSelection cuts all events in visual selection (copy then delete)
func HandleCutSelection(m Model, count int) (tea.Model, tea.Cmd) {
	// IMPORTANT: Get events BEFORE changing mode - visual selection function checks IsVisualMode
	selectedEvents := week_view_grid.GetEventsInVisualSelectionByTime(&m.WeekModel, m.EventManager)

	// Now exit visual mode
	m.VimState.SetMode(types.ModeNormal)
	m.WeekModel.IsVisualMode = false

	if len(selectedEvents) > 0 {
		// Convert SafeEvents to ical.Components and check for recurring events
		var events []*ical.Event
		var hasRecurringEvents bool
		for _, event := range selectedEvents {
			// Check if any selected event is recurring
			if util.IsRecurring(event) {
				hasRecurringEvents = true
				continue // Skip recurring events
			}
			events = append(events, event)
		}

		// Show error if any recurring events were in selection
		if hasRecurringEvents {
			util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, fmt.Errorf("unsupported operation"), "Copying recurring events")
		}

		// Cut only non-recurring events
		if len(events) > 0 {
			// Copy first
			err := m.EventManager.CopyEvents(events)
			util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Failed to copy events")
			if err == nil {
				// Then delete all events as a batch
				err, cmd := m.EventManager.DeleteEvents(events)
				if err != nil {
					util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Failed to delete events")
				} else if cmd != nil {
					return m, cmd
				}
			}
		}
	}
	return m, nil
}

func HandleDeleteSelection(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager
	vimState := m.VimState

	selectedEvents := week_view_grid.GetEventsInVisualSelectionByTime(weekModel, eventManager)
	if len(selectedEvents) > 0 {
		// Convert SafeEvents to ical.Components
		var components []*ical.Event
		for _, event := range selectedEvents {
			components = append(components, event)
		}
		// Delete all events as a batch
		err, cmd := eventManager.DeleteEvents(components)
		if err != nil {
			util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to delete events")
		} else if cmd != nil {
			return m, cmd
		}
	}
	vimState.SetMode(types.ModeNormal)
	weekModel.IsVisualMode = false
	return m, nil
}

// HandleSwapSelectionEnds swaps cursor and anchor positions
func HandleSwapSelectionEnds(m Model, count int) (tea.Model, tea.Cmd) {
	currentCursor := m.WeekModel.Cursor
	currentAnchor := m.WeekModel.VisualAnchor
	currentCursorWeek := m.WeekModel.CurrentlyViewedWeek
	currentAnchorWeek := m.WeekModel.VisualAnchorWeek

	// Swap the positions and weeks
	m.WeekModel.Cursor = currentAnchor
	m.WeekModel.VisualAnchor = currentCursor
	m.WeekModel.CurrentlyViewedWeek = currentAnchorWeek
	m.WeekModel.VisualAnchorWeek = currentCursorWeek
	m.VimState.SetVisualAnchor(currentCursor)
	return m, nil
}
