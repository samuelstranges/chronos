package app

import (
	"time"
	
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/keybinds_hints"
	"github.com/samuelstranges/chronos/navigation"
	// forms "github.com/samuelstranges/chronos/popup_forms" // REMOVED
	"github.com/samuelstranges/chronos/popup"
	"github.com/samuelstranges/chronos/search_parser"
	statusbar "github.com/samuelstranges/chronos/status_bar"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	week "github.com/samuelstranges/chronos/week_view"
)

// View renders the complete application view
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	// Build the base view: week + status bar
	baseView := m.renderBaseView()

	// Apply overlays in priority order (highest to lowest priority)
	return m.applyOverlays(baseView)
}

// renderBaseView creates the base week view with status bar
func (m Model) renderBaseView() string {
	searchResults := m.calculateSearchResults()
	weekViewStr := week.RenderWeekView(m.WeekModel, searchResults)

	// Always render status bar
	statusBarStr := statusbar.RenderStatusBar(&m.WeekModel, m.VimState, m.WeekModel.Width, searchResults)

	// Combine week view with status bar as the base view
	return lipgloss.JoinVertical(lipgloss.Left, weekViewStr, statusBarStr)
}

// applyOverlays applies overlays in priority order, returning final composed view
func (m Model) applyOverlays(baseView string) string {
	currentView := baseView

	// 1. KEY HINTS (highest priority - can override everything)
	if m.KeyHints.IsVisible() {
		bindings := m.KeyHints.GetBindingsForMode(m.VimState.GetMode())
		if len(bindings) > 0 {
			modeInfo := m.VimState.GetMode().GetModeInfo()
			keyHintContent := keybinds_hints.RenderKeyhints(bindings, m.WeekModel.Width, m.KeyHints.IsShowingAll(), modeInfo, m.WeekModel.CachedBackgroundColor)
			if keyHintContent != "" {
				return m.renderBottomOverlay(currentView, keyHintContent)
			}
		}
	}

	// 2. NEW POPUP SYSTEM (higher priority than old forms)
	if popup.HasPopup() {
		// Create render context with theme colors
		ctx := popup.PopupRenderContext{
			Width:           m.WeekModel.Width,
			Height:          m.WeekModel.Height,
			BackgroundColor: m.WeekModel.CachedBackgroundColor,
			ForegroundColor: "#ffffff", // White text
			BorderColor:     "62",      // Blue border (matches existing style)
		}

		popupContent := popup.RenderPopup(ctx)
		if popupContent != "" {
			return m.renderCenterOverlay(currentView, popupContent)
		}
	}

	// OLD FORM SYSTEM REMOVED

	// OLD POPUP OVERLAYS REMOVED - Now using popup_system exclusively

	// No overlays - return base view with proper height
	return m.renderFinalView(currentView)
}

// renderCenterOverlay renders an overlay in the center of the screen
func (m Model) renderCenterOverlay(baseView, overlayContent string) string {
	return week.RenderWithOverlay(
		baseView,
		overlayContent,
		m.WeekModel.Width,
		m.WeekModel.Height,
		0, // overlay package handles positioning
	)
}

// renderBottomOverlay renders an overlay at the bottom of the screen
func (m Model) renderBottomOverlay(baseView, overlayContent string) string {
	return week.RenderWithBottomOverlay(
		baseView,
		overlayContent,
		m.WeekModel.Width,
		m.WeekModel.Height,
		0, // overlay package handles positioning
	)
}

// OLD FORM SYSTEM REMOVED - replaced by popup_better_forms

// renderFinalView applies final styling to the base view
func (m Model) renderFinalView(baseView string) string {
	return lipgloss.NewStyle().
		Height(m.WeekModel.Height).
		AlignVertical(lipgloss.Top).
		Render(baseView)
}

// TODO: THIS METHOD DONT BELON GHEERE
// getActiveSearchInput returns the current search input (active or locked) and whether any search is active
func (m Model) getActiveSearchInput() (string, bool) {
	if m.WeekModel.SearchActive {
		return m.WeekModel.SearchInput, true
	}
	if m.WeekModel.SearchLocked {
		return m.WeekModel.LockedSearchInput, true
	}
	return "", false
}

// calculateSearchResults computes search results if search is active or locked
func (m Model) calculateSearchResults() []*types.EventInstance {
	searchInput, hasSearch := m.getActiveSearchInput()
	if !hasSearch {
		return nil
	}

	if searchInput == "" {
		// Search is active but no input yet - return empty results to grey out all events
		return []*types.EventInstance{}
	}

	field, phrase, valid := search_parser.ParseSearchInput(searchInput)
	if !valid {
		// Invalid search syntax - return empty results to grey out all events
		return []*types.EventInstance{}
	}

	if phrase == "" {
		// No phrase yet - return empty results to grey out all events
		return []*types.EventInstance{}
	}

	// Get matching events for view highlighting using navigation criteria
	criteria := navigation.SearchCriteria{Field: field, Phrase: phrase}
	
	// Get events in a small range around current week for highlighting (1 week either side)
	searchRange := 7 * 24 * time.Hour // 1 week
	currentWeekStart := m.WeekModel.CurrentlyViewedWeek
	searchStart := currentWeekStart.Add(-searchRange)
	searchEnd := currentWeekStart.Add(searchRange)
	
	instances, err := m.EventManager.GetEventsForDateRange(nil, searchStart, searchEnd)
	if err != nil {
		return []*types.EventInstance{}
	}
	
	// Filter to matching events
	var matchingInstances []*types.EventInstance
	for _, instance := range instances {
		// Skip all-day events
		if util.IsAllDayEvent(instance.OriginalEvent) {
			continue
		}
		
		// Create positioned event to check criteria
		positionedEvent := types.PositionedEvent{
			Instance:    instance,
			IsStartCell: true, // For search highlighting, we only care about starts
			IsEndCell:   false,
		}
		
		if criteria.Matches(positionedEvent) {
			matchingInstances = append(matchingInstances, instance)
		}
	}
	
	return matchingInstances
}
