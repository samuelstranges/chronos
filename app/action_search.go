package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/navigation"
	"github.com/samuelstranges/chronos/search_parser"
	types "github.com/samuelstranges/chronos/types"
)

// HandleEnterSearchMode enters search mode (called from app layer with vimState access)
func HandleEnterSearchMode(m Model, count int) (tea.Model, tea.Cmd) {
	// Clear any locked search when starting new search
	m.WeekModel.SearchLocked = false
	m.WeekModel.LockedSearchInput = ""
	m.WeekModel.CurrentSearchIdx = -1

	// Set search state
	m.WeekModel.SearchActive = true
	m.WeekModel.SearchInput = "/"

	m.VimState.SetMode(types.ModeSearch)

	return m, nil
}

func HandleExitSearchMode(m Model, count int) (tea.Model, tea.Cmd) {
	// Clear all search state (both active and locked)
	m.WeekModel.SearchActive = false
	m.WeekModel.SearchInput = ""
	m.WeekModel.SearchLocked = false
	m.WeekModel.LockedSearchInput = ""
	m.WeekModel.CurrentSearchIdx = -1
	
	m.VimState.SetMode(types.ModeNormal)
	return m, nil
}

func HandleExecuteSearch(m Model, count int) (tea.Model, tea.Cmd) {
	// Parse the search input to validate
	_, phrase, valid := search_parser.ParseSearchInput(m.WeekModel.SearchInput)
	if valid && phrase != "" {
		// Lock the search
		m.WeekModel.SearchLocked = true
		m.WeekModel.LockedSearchInput = m.WeekModel.SearchInput
		m.WeekModel.CurrentSearchIdx = -1 // No position selected yet
		// Keep SearchActive = true and stay in search mode for n/N navigation
	}
	return m, nil
}

// HandleNavigateToNextResult moves cursor to next search result (actual implementation)
func HandleNavigateToNextResult(m Model, count int) (tea.Model, tea.Cmd) {
	// Only work if search is locked
	if !m.WeekModel.SearchLocked {
		return m, nil
	}

	// Parse the locked search
	field, phrase, valid := search_parser.ParseSearchInput(m.WeekModel.LockedSearchInput)
	if !valid || phrase == "" {
		return m, nil
	}

	// Use unified navigation
	criteria := navigation.SearchCriteria{Field: field, Phrase: phrase}
	if navigation.Navigate(&m.WeekModel, m.EventManager, navigation.Forward, criteria) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

// HandleNavigateToPreviousResult moves cursor to previous search result (actual implementation)
func HandleNavigateToPreviousResult(m Model, count int) (tea.Model, tea.Cmd) {
	// Only work if search is locked
	if !m.WeekModel.SearchLocked {
		return m, nil
	}

	// Parse the locked search
	field, phrase, valid := search_parser.ParseSearchInput(m.WeekModel.LockedSearchInput)
	if !valid || phrase == "" {
		return m, nil
	}

	// Use unified navigation
	criteria := navigation.SearchCriteria{Field: field, Phrase: phrase}
	if navigation.Navigate(&m.WeekModel, m.EventManager, navigation.Backward, criteria) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

func HandleClearLockedSearch(m Model, count int) (tea.Model, tea.Cmd) {
	// Clear only the locked search, keep other state
	m.WeekModel.SearchLocked = false
	m.WeekModel.LockedSearchInput = ""
	m.WeekModel.CurrentSearchIdx = -1
	return m, nil
}