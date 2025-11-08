// /Users/computer/Documents/Git/new_chronos/app/handleQuit.go

package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	keybinds_hints "github.com/samuelstranges/chronos/keybinds_hints"
	"github.com/samuelstranges/chronos/popup"
	types "github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// HandleQuitApplication quits the application
func HandleQuit(m Model, count int) (tea.Model, tea.Cmd) {
	m.Quitting = true
	return m, tea.Sequence(tea.ClearScreen, tea.ExitAltScreen, tea.Quit)
}

// /Users/computer/Documents/Git/new_chronos/app/handleGenericForm.go

// HandleCloseForm closes the currently visible form
func HandleCloseForm(m Model, count int) (tea.Model, tea.Cmd) {
	showForm := &m.ShowForm
	form := &m.Form
	*showForm = false
	*form = nil
	return m, nil
}

// /Users/computer/Documents/Git/new_chronos/app/handleLayers.go

// HandleToggleAgendaLayer shows the agenda picker popup for the current cursor day
func HandleToggleAgendaLayer(m Model, count int) (tea.Model, tea.Cmd) {
	dayIndex := m.WeekModel.Cursor.Day
	dayName := getDayName(dayIndex)

	// Get events for the day under cursor
	dayEvents, err := m.EventManager.GetEventsForDayUnderCursor(&m.WeekModel)
	if err != nil {
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to get events")
		return m, nil
	}

	// Show the agenda picker popup
	popup.ShowAgendaPicker(dayEvents, dayIndex, dayName, func(action popup.AgendaPickerAction, event *ical.Event) tea.Cmd {
		return handleAgendaPickerAction(m, action, event, dayEvents)
	})

	return m, nil
}

// getDayName returns the day name for a day index
func getDayName(dayIndex int) string {
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if dayIndex >= 0 && dayIndex < len(dayNames) {
		return dayNames[dayIndex]
	}
	return "Unknown"
}

func HandleToggleAllDayGrid(m Model, count int) (tea.Model, tea.Cmd) {
	m.WeekModel.ShowAllDayGrid = !m.WeekModel.ShowAllDayGrid
	return m, nil
}

func HandleToggleEventTextColor(m Model, count int) (tea.Model, tea.Cmd) {
	m.WeekModel.EventTextBlack = !m.WeekModel.EventTextBlack
	return m, nil
}

// /Users/computer/Documents/Git/new_chronos/app/handleUndoRedo.go

// HandleUndoAction performs undo operation
func HandleUndoAction(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager

	// Undo operation (refresh happens automatically)
	err, cmd := eventManager.Undo()
	util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to undo")
	return m, cmd
}

// HandleRedoAction performs redo operation
func HandleRedoAction(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	eventManager := m.EventManager

	// Redo operation (refresh happens automatically)
	err, cmd := eventManager.Redo()
	util.ShowIfError(weekModel, types.ErrorDisplaySeconds, err, "Failed to redo")
	return m, cmd
}

// /Users/computer/Documents/Git/new_chronos/app/handleHints.go

// HandleToggleHints toggles keyhints visibility based on current state
func HandleToggleHints(m Model, count int) (tea.Model, tea.Cmd) {
	keyHints := m.KeyHints
	vimMode := m.VimState.GetMode()
	statusText := m.VimState.GetStatusText()
	if keyHints.IsVisible() {
		// If visible, toggle show all or hide
		if keyHints.IsShowingAll() {
			return HandleHideHints(m, count)
		}
		keyHints.ToggleShowAll()
		return m, nil
	} else {
		// If not visible, show hints for current vim state
		keyHints.SetMode(keybinds_hints.ConvertVimMode(int(vimMode)))
		keyHints.ShowForVimState(statusText)
		return m, nil
	}
}

// HandleHideHints hides key hints and cancels any pending timeout
func HandleHideHints(m Model, count int) (tea.Model, tea.Cmd) {
	keyHints := m.KeyHints
	waitingForTimeout := &m.WaitingForKeyTimeout
	*waitingForTimeout = false // Cancel any pending timeout
	keyHints.Hide()
	return m, nil
}

// HandleUpdateHintsForSequence immediately updates hints if visible, or starts timeout if not
func HandleUpdateHintsForSequence(m Model, count int) (tea.Model, tea.Cmd) {
	keyHints := m.KeyHints
	vimMode := m.VimState.GetMode()
	statusText := m.VimState.GetStatusText()
	if keyHints.IsVisible() {
		// Hints are already visible, update them immediately
		keyHints.SetMode(keybinds_hints.ConvertVimMode(int(vimMode)))
		keyHints.ShowForVimState(statusText)
		return m, nil
	} else {
		// Hints not visible, start timeout to show them
		return HandleStartHintTimeout(m, count)
	}
}

// HandleStartHintTimeout starts timeout for showing key hints
func HandleStartHintTimeout(m Model, count int) (tea.Model, tea.Cmd) {
	waitingForTimeout := &m.WaitingForKeyTimeout
	if !*waitingForTimeout {
		*waitingForTimeout = true
		return m, keybinds_hints.CreateKeyHintTimeout()
	}
	return m, nil
}
