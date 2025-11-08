package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// HandleZoomIn implements zoom_in command
func HandleZoomIn(m Model, count int) (tea.Model, tea.Cmd) {
	for i := 0; i < count; i++ {
		m.WeekModel = week_view_grid.ZoomIn(m.WeekModel)
	}
	return m, func() tea.Msg { return types.RefreshMsg{} }
}

// HandleZoomOut implements zoom_out command
func HandleZoomOut(m Model, count int) (tea.Model, tea.Cmd) {
	for i := 0; i < count; i++ {
		m.WeekModel = week_view_grid.ZoomOut(m.WeekModel)
	}
	return m, func() tea.Msg { return types.RefreshMsg{} }
}