package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/navigation"
	"github.com/samuelstranges/chronos/types"
)

// HandleNavigateToNextEvent handles w command - finds and navigates to next event across all weeks
func HandleNavigateToNextEvent(m Model, count int) (tea.Model, tea.Cmd) {
	if navigation.Navigate(&m.WeekModel, m.EventManager, navigation.Forward, navigation.EventStartCriteria{}) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

func HandleNavigateToPreviousEvent(m Model, count int) (tea.Model, tea.Cmd) {
	if navigation.Navigate(&m.WeekModel, m.EventManager, navigation.Backward, navigation.EventStartCriteria{}) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

// HandleNavigateToEventEnd handles e command - vim-like 'e' behavior for event ends
func HandleNavigateToEventEnd(m Model, count int) (tea.Model, tea.Cmd) {
	if navigation.Navigate(&m.WeekModel, m.EventManager, navigation.Forward, navigation.EventEndCriteria{}) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

