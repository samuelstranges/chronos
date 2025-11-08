package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/huh/v2"
	"github.com/samuelstranges/chronos/ical_crud"
	"github.com/samuelstranges/chronos/keybinds_hints"
	keyhints "github.com/samuelstranges/chronos/keybinds_hints"
	"github.com/samuelstranges/chronos/popup"
	"github.com/samuelstranges/chronos/popup_better_forms"
	"github.com/samuelstranges/chronos/search_parser"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_all_day_grid"
	"github.com/samuelstranges/chronos/week_view_grid"
	"github.com/samuelstranges/chronos/week_view_shared"
)

// Update handles all application state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size messages first (don't pass to form)
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.WeekModel.Width = msg.Width
		m.WeekModel.Height = msg.Height
		m.WeekModel.CachedCellWidth, m.WeekModel.CachedContentWidth = week_view_shared.CalculateCellWidth(msg.Width)
		return m, nil
	}

	// NEW POPUP SYSTEM: Popup gets first shot at ALL input when active
	if popup.HasPopup() {
		// Special handling for FormPopup - pass raw tea.Msg like old system
		if currentPopup := popup.GlobalPopupManager.GetCurrent(); currentPopup != nil {
			if formPopup, ok := currentPopup.(*popup_better_forms.FormPopup); ok {
				return m.handleFormPopupUpdate(msg, formPopup)
			}
		}

		// Regular popup handling for non-form popups
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			cmd := popup.HandlePopupKey(keyMsg.String())
			return m, cmd
		}
	}

	// OLD SYSTEM: Still handling individual popups temporarily for migration
	// TODO: Remove these once all popups are migrated to popup_system

	// OLD FORM SYSTEM REMOVED - handled by popup_better_forms now

	// OLD POPUP OVERLAY HANDLING REMOVED - Now handled by popup_system

	// If search is active, handle search input directly (like forms)
	if m.WeekModel.SearchActive {
		return m.handleSearchUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyUpdate(msg)
	case types.VimActionMsg:
		return handleVimAction(m, msg)
	case types.VimErrorMsg:
		// Handle vim errors by displaying them in the status bar
		util.ShowIfError(&m.WeekModel, 5, fmt.Errorf(msg.Error), "Vim error")
		return m, nil
	case types.RefreshMsg:
		// Handle refresh requests from commands
		err := refreshEventGridsWithEventManager(&m.WeekModel, m.EventManager)
		util.ShowIfError(&m.WeekModel, 5, err, "Failed to refresh events")
		return m, nil
	case types.BackgroundSyncMsg:
		// Handle background calendar sync (async)
		return m, performBackgroundSync(m.EventManager, m.SyncInterval)
	case types.RecurringDialogChoiceMsg:
		return handleRecurringDialogChoice(m, msg)
	case keybinds_hints.KeyHintTimeoutMsg:
		return m, keybinds_hints.HandleKeyHintTimeout(m.KeyHints, m.VimState.GetMode(), m.VimState.GetStatusText(), &m.WaitingForKeyTimeout)
	}
	return m, nil
}

// clearSearchState clears all search-related state
func (m *Model) clearSearchState() {
	m.WeekModel.SearchActive = false
	m.WeekModel.SearchInput = ""
	m.WeekModel.SearchLocked = false
	m.WeekModel.LockedSearchInput = ""
	m.WeekModel.CurrentSearchIdx = -1
}

// handleSearchUpdate processes search input (similar to form handling)
func (m Model) handleSearchUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Only handle key messages in search mode
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	keyStr := keyMsg.String()

	switch keyStr {
	case "esc", "ctrl+c":
		// Exit search mode and clear all search state
		m.clearSearchState()
		m.VimState.SetMode(types.ModeNormal)
		return m, nil
	case "enter":
		// Lock search but stay in search mode
		_, phrase, valid := search_parser.ParseSearchInput(m.WeekModel.SearchInput)
		if valid && phrase != "" {
			// Lock the search
			m.WeekModel.SearchLocked = true
			m.WeekModel.LockedSearchInput = m.WeekModel.SearchInput
			m.WeekModel.CurrentSearchIdx = -1 // No position selected yet
			// Set SearchActive = false to prevent further typing but stay in search mode for n/N navigation
			m.WeekModel.SearchActive = false
		}
		return m, nil
	case "backspace":
		// Remove last character
		if len(m.WeekModel.SearchInput) > 1 { // Keep at least the "/"
			m.WeekModel.SearchInput = m.WeekModel.SearchInput[:len(m.WeekModel.SearchInput)-1]
		}
		return m, nil
	default:
		// Add regular characters to search input
		// Only allow printable characters
		if len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
			m.WeekModel.SearchInput += keyStr
		}
		return m, nil
	}
}

// handleKeyUpdate processes keyboard input
func (m Model) handleKeyUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Observer pattern: Handle backspace for keyhints navigation (non-intercepting)
	if keyStr == "backspace" && m.KeyHints.IsVisible() {
		keyhints.HandleNavigateHintsBack(m.KeyHints)
	}

	// Delegate ALL key handling to vim system
	if vimActionMsg := m.VimState.HandleKey(msg, m.VimRegistry, m.ShowForm); vimActionMsg.Action != "" {
		// Store old week for comparison
		oldWeek := m.WeekModel.CurrentlyViewedWeek

		// Command was completed - handle via vim action system
		m.WaitingForKeyTimeout = false
		model, cmd := handleVimAction(m, vimActionMsg)
		m = model.(Model)

		// Check if week changed and refresh events if needed
		if !m.WeekModel.CurrentlyViewedWeek.Equal(oldWeek) {
			refreshCmd := func() tea.Msg { return types.RefreshMsg{} }
			if cmd != nil {
				return m, tea.Sequence(cmd, refreshCmd)
			}
			return m, refreshCmd
		}

		return m, cmd
	}

	return m, nil
}

// refreshEventGrids recalculates what should be displayed on screen
//
// ARCHITECTURE: Display Layer Implementation (Private)
// ===================================================
//
// This is the "response" side of the observer pattern:
//   EventManager signals → "Data changed, here's current state"
//   This function responds → "Got it, updating display now"
//
// RESPONSIBILITIES:
//   - Recalculate event grids for all zoom levels
//   - Update all-day event grids
//   - Clamp cursor position to valid bounds
//   - Single source of truth for display updates
//
// This is intentionally PRIVATE to enforce proper architecture:
//   - External code MUST go through EventManager callback system
//   - No direct calls that bypass the clean message passing
//   - Ensures consistent refresh behavior across all operations
//

// refreshEventGridsWithEventManager recalculates the display grids using EventInstances directly
func refreshEventGridsWithEventManager(weekModel *types.WeekModel, eventManager *ical_crud.EventManager) error {
	// Recalculate all zoom levels using EventInstances directly
	zoomLevels := []types.ZoomLevel{types.Zoom30Min, types.Zoom15Min, types.Zoom5Min, types.Zoom1Min}
	for i, zoomLevel := range zoomLevels {
		weekLayout, layoutErr := week_view_grid.BuildTimedGrid(eventManager, weekModel, zoomLevel)
		if layoutErr != nil {
			// Use empty layout on error
			weekLayout = types.WeekEventGrid{StartDate: weekModel.CurrentlyViewedWeek}
		}
		weekModel.WeekEventGrids[i] = weekLayout
	}

	// Calculate all-day event grid using only all-day events from EventManager
	allDayEventInstances, err := eventManager.GetAllDayEventsForWeek(weekModel)
	if err != nil {
		// Fall back to empty grids on error
		return initializeEmptyGrids(weekModel)
	}

	allDayGrid, allDayCounts, err := week_view_all_day_grid.CalculateAllDayEventGridFromInstances(allDayEventInstances, weekModel.CurrentlyViewedWeek)
	if err != nil {
		// Use empty grid on error
		allDayGrid = types.AllDayEventGrid{StartDate: weekModel.CurrentlyViewedWeek}
		allDayCounts = []int{0, 0, 0, 0, 0, 0, 0}
	}
	weekModel.AllDayEventGrid = allDayGrid
	weekModel.AllDayEventCounts = allDayCounts

	// Ensure cursor column is still valid after display refresh
	weekModel.Cursor = week_view_grid.ClampCursorColumn(weekModel, weekModel.Cursor)

	return nil
}

// initializeEmptyGrids creates empty grids when event loading fails
func initializeEmptyGrids(weekModel *types.WeekModel) error {
	zoomLevels := []types.ZoomLevel{types.Zoom30Min, types.Zoom15Min, types.Zoom5Min, types.Zoom1Min}
	for i := range zoomLevels {
		weekModel.WeekEventGrids[i] = types.WeekEventGrid{StartDate: weekModel.CurrentlyViewedWeek}
	}

	weekModel.AllDayEventGrid = types.AllDayEventGrid{StartDate: weekModel.CurrentlyViewedWeek}
	weekModel.AllDayEventCounts = []int{0, 0, 0, 0, 0, 0, 0}

	return nil
}

// handleFormPopupUpdate handles form updates using raw tea.Msg (like old system)
func (m Model) handleFormPopupUpdate(msg tea.Msg, formPopup *popup_better_forms.FormPopup) (tea.Model, tea.Cmd) {
	// Handle ESC to cancel form
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		popup.ClosePopup()
		return m, nil
	}

	// Update form with raw message (like old system did)
	form := formPopup.GetForm()
	if form != nil {
		updatedForm, cmd := form.Update(msg)
		if newForm, ok := updatedForm.(*huh.Form); ok {
			formPopup.SetForm(newForm)

			// Check if form is completed
			if formPopup.IsCompleted() {
				popup.ClosePopup()
				// Execute the submit callback if it exists
				if submitCallback := formPopup.OnSubmit(); submitCallback != nil {
					return m, submitCallback()
				}
			}
		}
		return m, cmd
	}

	return m, nil
}

// performBackgroundSync reloads calendars from server in background (async)
// Returns a tea.Cmd that schedules the next sync tick
func performBackgroundSync(eventManager *ical_crud.EventManager, syncInterval int) tea.Cmd {
	return tea.Batch(
		// Async reload from server
		func() tea.Msg {
			if err := eventManager.ReloadFromStorage(); err != nil {
				return types.VimErrorMsg{Error: fmt.Sprintf("Background sync failed: %v", err)}
			}
			return types.RefreshMsg{} // Refresh UI after successful sync
		},
		// Schedule next sync
		tea.Tick(time.Duration(syncInterval)*time.Second, func(t time.Time) tea.Msg {
			return types.BackgroundSyncMsg{}
		}),
	)
}

