package week_view

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_all_day_grid"
	"github.com/samuelstranges/chronos/week_view_grid"
)

func getUnreservedLines(m types.WeekModel) int {
	// Calculate reserved lines dynamically based on whether all-day grid is shown
	baseHeaderLines := types.MonthHeaderLines + types.DayTitleLines
	allDayLines := 0
	if m.ShowAllDayGrid {
		allDayLines = types.AllDayCountRow + types.AllDayCells + types.AllDaySeperator
	}

	reservedLines := baseHeaderLines + allDayLines + types.FooterLines
	unreserved := m.Height - reservedLines
	return max(types.MinVisibleCellsPerDay, unreserved)
}

func getNumOfPossibleVisibleCells(m types.WeekModel) int {
	maxCells := util.GetCellsPerDay(m.CurrentZoom)
	possibleRows := getUnreservedLines(m)
	return min(maxCells, possibleRows)
}

// PUBLIC METHODS
func RenderWeekView(m types.WeekModel, searchResults []*types.EventInstance) string {
	// Render month/year header
	monthYearHeader := renderMonthYearHeader(m)

	dayHeader := renderDaysHeadersHorizontally(m, m.CurrentlyViewedWeek)

	// Conditionally render all-day grid section
	var allDayRows string
	if m.ShowAllDayGrid {
		allDayRows = week_view_all_day_grid.RenderAllDayEventSection(m, m.AllDayEventCounts)
	}

	// availableRows := m.getUnreservedLines()
	numOfPossibleVisibleCells := getNumOfPossibleVisibleCells(m)
	maxCells := util.GetCellsPerDay(m.CurrentZoom)

	// Get starting row
	idealStartingRow := m.Cursor.Cell - numOfPossibleVisibleCells/types.HalfDivisor
	maxStartingRow := maxCells - numOfPossibleVisibleCells
	startRow := max(0, min(idealStartingRow, maxStartingRow)) // cant be negative or past end of day

	var dayColumns []string
	timeColumn := week_view_grid.RenderTimeColumn(m, startRow, numOfPossibleVisibleCells, m.CurrentZoom)
	dayColumns = append(dayColumns, timeColumn)

	for day := range types.DaysPerWeek {
		dayCol := week_view_grid.RenderDayColumn(m, day, startRow, numOfPossibleVisibleCells, searchResults)
		dayColumns = append(dayColumns, dayCol)
	}

	weekGrid := lipgloss.JoinHorizontal(lipgloss.Top, dayColumns...)

	// Combine month/year header with week grid
	// Build sections list conditionally
	sections := []string{monthYearHeader, dayHeader}
	if m.ShowAllDayGrid {
		sections = append(sections, allDayRows)
	}
	sections = append(sections, weekGrid)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

var (
	borderedOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1)

	fullWidthOverlayStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1e1e2e"))
)

// RenderWithOverlay renders a base view with an optional overlay on top
func RenderWithOverlay(baseView, overlayView string, width, height int, overlayHeight int) string {
	if overlayView == "" {
		return baseView
	}

	// Add a border around the overlay
	borderedOverlay := borderedOverlayStyle.Render(overlayView)

	// Create layers for the base and overlay
	baseLayer := lipgloss.NewLayer(baseView)
	overlayLayer := lipgloss.NewLayer(borderedOverlay)

	// Calculate actual dimensions of the bordered overlay
	actualWidth := lipgloss.Width(borderedOverlay)
	actualHeight := lipgloss.Height(borderedOverlay)

	// Center the overlay in the terminal window
	centerX, centerY := CenterContent(actualWidth, actualHeight, width, height)

	// Create canvas with positioned layers
	canvas := lipgloss.NewCanvas(
		baseLayer.X(0).Y(0).Z(0),
		overlayLayer.X(centerX).Y(centerY).Z(1),
	)

	return canvas.Render()
}

// RenderWithBottomOverlay renders a base view with an overlay positioned at the bottom
func RenderWithBottomOverlay(baseView, overlayView string, width, height int, overlayHeight int) string {
	if overlayView == "" {
		return baseView
	}

	// Make overlay full width without border - just background styling
	fullWidthOverlay := fullWidthOverlayStyle.Width(width).Render(overlayView)
	actualHeight := lipgloss.Height(fullWidthOverlay)

	// Create layers for the base and overlay
	baseLayer := lipgloss.NewLayer(baseView)
	overlayLayer := lipgloss.NewLayer(fullWidthOverlay)

	// Position at bottom, full width
	bottomY := height - actualHeight - 1 // - 1 to show statusbar
	bottomY = max(bottomY, 0)            // Ensure we don't go negative

	// Create canvas with positioned layers
	canvas := lipgloss.NewCanvas(
		baseLayer.X(0).Y(0).Z(0),
		overlayLayer.X(0).Y(bottomY).Z(1),
	)

	return canvas.Render()
}
