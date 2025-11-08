package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// Movement function types for functional composition
type (
	moveFunc     func(*types.WeekModel) bool
	boundaryFunc func(*types.WeekModel)
)

// Helper functions for functional movement
func dec(val *int) bool {
	if *val > 0 {
		*val--
		return true
	}
	return false
}

func inc(val *int, maxVal int) bool {
	if *val < maxVal {
		*val++
		return true
	}
	return false
}

// Generic movement with injected behavior
func moveWithBehavior(weekModel *types.WeekModel, count int,
	primaryMove moveFunc,
	boundaryMove moveFunc,
	onBoundary boundaryFunc,
) tea.Cmd {
	for range count {
		// Try primary movement first
		if !primaryMove(weekModel) {
			// If primary fails, try boundary movement
			if boundaryMove(weekModel) {
				onBoundary(weekModel)
			}
		}
	}

	weekModel.Cursor = week_view_grid.ClampCursorColumn(weekModel, weekModel.Cursor)
	return nil
}

// HandleMoveLeft implements h command
func HandleMoveLeft(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel

	// In visual mode, always use simple day movement (no column logic)
	if weekModel.IsVisualMode || count > 1 {
		// Simple day movement for count > 1
		moveWithBehavior(weekModel, count,
			func(wm *types.WeekModel) bool { return dec(&wm.Cursor.Day) },
			func(wm *types.WeekModel) bool {
				// Hit week boundary - go to previous week
				week_view_grid.JumpToPreviousWeek(wm)
				wm.Cursor.Day = types.MaxDayIndex // Move to last day of new week
				return true
			},
			func(wm *types.WeekModel) { wm.Cursor.EventColumn = 0 },
		)
	} else {
		// Smart column/day movement for count == 1
		moveWithBehavior(weekModel, count,
			func(wm *types.WeekModel) bool {
				maxCols := week_view_grid.GetCellColumnCount(wm, wm.Cursor.Day, wm.Cursor.Cell)
				return maxCols > 1 && dec(&wm.Cursor.EventColumn)
			},
			func(wm *types.WeekModel) bool {
				if dec(&wm.Cursor.Day) {
					return true
				}
				// Hit week boundary - go to previous week
				week_view_grid.JumpToPreviousWeek(wm)
				wm.Cursor.Day = types.MaxDayIndex // Move to last day of new week
				return true
			},
			func(wm *types.WeekModel) {
				maxCols := week_view_grid.GetCellColumnCount(wm, wm.Cursor.Day, wm.Cursor.Cell)
				if maxCols > 0 {
					wm.Cursor.EventColumn = maxCols - 1
				} else {
					wm.Cursor.EventColumn = 0
				}
			},
		)
	}
	return m, nil
}

// HandleMoveRight implements l command
func HandleMoveRight(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel

	// In visual mode, always use simple day movement (no column logic)
	if weekModel.IsVisualMode || count > 1 {
		// Simple day movement for count > 1
		moveWithBehavior(weekModel, count,
			func(wm *types.WeekModel) bool { return inc(&wm.Cursor.Day, types.MaxDayIndex) },
			func(wm *types.WeekModel) bool {
				// Hit week boundary - go to next week
				week_view_grid.JumpToNextWeek(wm)
				wm.Cursor.Day = 0 // Move to first day of new week
				return true
			},
			func(wm *types.WeekModel) { wm.Cursor.EventColumn = 0 },
		)
	} else {
		// Smart column/day movement for count == 1
		moveWithBehavior(weekModel, count,
			func(wm *types.WeekModel) bool {
				maxCols := week_view_grid.GetCellColumnCount(wm, wm.Cursor.Day, wm.Cursor.Cell)
				return maxCols > 1 && inc(&wm.Cursor.EventColumn, maxCols-1)
			},
			func(wm *types.WeekModel) bool {
				if inc(&wm.Cursor.Day, types.MaxDayIndex) {
					return true
				}
				// Hit week boundary - go to next week
				week_view_grid.JumpToNextWeek(wm)
				wm.Cursor.Day = 0 // Move to first day of new week
				return true
			},
			func(wm *types.WeekModel) { wm.Cursor.EventColumn = 0 },
		)
	}
	return m, nil
}

// MoveDown implements j command
func HandleMoveDown(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	maxCells := util.GetCellsPerDay(weekModel.CurrentZoom)

	moveWithBehavior(weekModel, count,
		func(m *types.WeekModel) bool { return inc(&m.Cursor.Cell, maxCells-1) },
		func(m *types.WeekModel) bool {
			if inc(&m.Cursor.Day, types.MaxDayIndex) {
				return true
			}
			// Hit week boundary - go to next week
			week_view_grid.JumpToNextWeek(m)
			m.Cursor.Day = 0 // Move to first day of new week
			return true
		},
		func(m *types.WeekModel) {
			m.Cursor.Cell = 0
			m.Cursor.EventColumn = 0
		},
	)
	return m, nil
}

// MoveUp implements k command
func HandleMoveUp(m Model, count int) (tea.Model, tea.Cmd) {
	weekModel := &m.WeekModel
	maxCells := util.GetCellsPerDay(weekModel.CurrentZoom)

	moveWithBehavior(weekModel, count,
		func(m *types.WeekModel) bool { return dec(&m.Cursor.Cell) },
		func(m *types.WeekModel) bool {
			if dec(&m.Cursor.Day) {
				return true
			}
			// Hit week boundary - go to previous week
			week_view_grid.JumpToPreviousWeek(m)
			m.Cursor.Day = types.MaxDayIndex // Move to last day of new week
			return true
		},
		func(m *types.WeekModel) {
			m.Cursor.Cell = maxCells - 1
			m.Cursor.EventColumn = 0
		},
	)
	return m, nil
}

// HandleJumpToCurrentTime implements t command
func HandleJumpToCurrentTime(m Model, count int) (tea.Model, tea.Cmd) {
	// Jump commands ignore count
	week_view_grid.JumpToCurrentTime(&m.WeekModel)
	return m, func() tea.Msg { return types.RefreshMsg{} }
}

// HandleJumpToStartOfDay implements gg command
func HandleJumpToStartOfDay(m Model, count int) (tea.Model, tea.Cmd) {
	// Jump commands ignore count
	week_view_grid.JumpToStartOfDay(&m.WeekModel)
	return m, func() tea.Msg { return types.RefreshMsg{} }
}

// HandleJumpToEndOfDay implements G command
func HandleJumpToEndOfDay(m Model, count int) (tea.Model, tea.Cmd) {
	// Jump commands ignore count
	week_view_grid.JumpToEndOfDay(&m.WeekModel)
	return m, func() tea.Msg { return types.RefreshMsg{} }
}

// HandleJumpToFirstEventOfDay implements ^ command
func HandleJumpToFirstEventOfDay(m Model, count int) (tea.Model, tea.Cmd) {
	// Use grid-first navigation to find first event of current day
	if week_view_grid.FindFirstEventOfDay(&m.WeekModel) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

// HandleJumpToLastEventOfDay implements $ command
func HandleJumpToLastEventOfDay(m Model, count int) (tea.Model, tea.Cmd) {
	// Use grid-first navigation to find last event of current day
	if week_view_grid.FindLastEventOfDay(&m.WeekModel) {
		return m, func() tea.Msg { return types.RefreshMsg{} }
	}
	return m, nil
}

// Visual mode movement commands (different behavior)

// VisualMoveLeft implements h in visual mode
func VisualMoveLeft(weekModel *types.WeekModel, count int) tea.Cmd {
	// In visual mode, h/l should move by days and cross week boundaries
	moveWithBehavior(weekModel, count,
		func(wm *types.WeekModel) bool { return dec(&wm.Cursor.Day) },
		func(wm *types.WeekModel) bool {
			// Hit week boundary - go to previous week
			week_view_grid.JumpToPreviousWeek(wm)
			wm.Cursor.Day = types.MaxDayIndex // Move to last day of new week
			return true
		},
		func(wm *types.WeekModel) { wm.Cursor.EventColumn = 0 },
	)
	return nil
}

// VisualMoveRight implements l in visual mode
func VisualMoveRight(weekModel *types.WeekModel, count int) tea.Cmd {
	// In visual mode, h/l should move by days and cross week boundaries
	moveWithBehavior(weekModel, count,
		func(wm *types.WeekModel) bool { return inc(&wm.Cursor.Day, 6) },
		func(wm *types.WeekModel) bool {
			// Hit week boundary - go to next week
			week_view_grid.JumpToNextWeek(wm)
			wm.Cursor.Day = 0 // Move to first day of new week
			return true
		},
		func(wm *types.WeekModel) { wm.Cursor.EventColumn = 0 },
	)
	return nil
}

// VisualMoveDown implements j in visual mode
func VisualMoveDown(weekModel *types.WeekModel, count int) tea.Cmd {
	// In visual mode, j/k should only move by time slots (no day wrapping)
	maxCells := util.GetCellsPerDay(weekModel.CurrentZoom)

	cmd := moveWithBehavior(weekModel, count,
		func(m *types.WeekModel) bool { return inc(&m.Cursor.Cell, maxCells-1) },
		func(_ *types.WeekModel) bool { return false }, // No boundary wrapping to next day
		func(_ *types.WeekModel) { /* no boundary action needed */ },
	)

	return cmd
}

// VisualMoveUp implements k in visual mode
func VisualMoveUp(weekModel *types.WeekModel, count int) tea.Cmd {
	// In visual mode, j/k should only move by time slots (no day wrapping)
	moveWithBehavior(weekModel, count,
		func(wm *types.WeekModel) bool { return dec(&wm.Cursor.Cell) },
		func(_ *types.WeekModel) bool { return false }, // No boundary wrapping to previous day
		func(wm *types.WeekModel) { /* no boundary action needed */ },
	)
	return nil
}

// NavigateToPreviousWeek implements H command
func HandleNavigateToPreviousWeek(m Model, count int) (tea.Model, tea.Cmd) {
	for i := 0; i < count; i++ {
		week_view_grid.JumpToPreviousWeek(&m.WeekModel)
	}
	// Preserve cursor position (day, cell, eventColumn remain the same)
	return m, func() tea.Msg { return types.RefreshMsg{} }
}

// NavigateToNextWeek implements L command
func HandleNavigateToNextWeek(m Model, count int) (tea.Model, tea.Cmd) {
	for i := 0; i < count; i++ {
		week_view_grid.JumpToNextWeek(&m.WeekModel)
	}
	// Preserve cursor position (day, cell, eventColumn remain the same)
	return m, func() tea.Msg { return types.RefreshMsg{} }
}
