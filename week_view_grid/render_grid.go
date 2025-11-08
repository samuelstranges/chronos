package week_view_grid

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_shared"
)

// RenderTimeColumn renders the time column on the left side of the grid
func RenderTimeColumn(m types.WeekModel, startingCellIndex int, cellsToRender int, minsPerRow types.ZoomLevel) string {
	var cells []string

	for i := range cellsToRender {
		cellIndex := startingCellIndex + i // in case we want blanks at top

		// Build actual string
		minutes := cellIndex * int(minsPerRow)
		hour := minutes / types.MinsPerHour
		min := minutes % types.MinsPerHour
		timeStr := fmt.Sprintf(types.TimeFormat, hour, min)

		// 3 cases: cursor on same row; current time on same row; normal
		isCursorCell := cellIndex == m.Cursor.Cell
		currentTimeCell := util.TimeToCell(time.Now(), m.CurrentZoom)
		isCurrentTime := currentTimeCell == cellIndex

		style := week_view_shared.TimeStyle // default
		if isCurrentTime {
			style = week_view_shared.TimeStyleCurrentTimeCell
		}
		if isCursorCell { // highest priority
			style = week_view_shared.TimeStyleCursorCell
		}

		cells = append(cells, style.Render(timeStr))
	}

	return lipgloss.JoinVertical(lipgloss.Left, cells...)
}

// GetCellLayout retrieves the cell layout for a specific day and cell index
func GetCellLayout(m types.WeekModel, day, cellIndex int) types.CellLayout {
	zoomIndex := GetZoomIndex(m.CurrentZoom)

	// Get the layout for current zoom level
	if zoomIndex >= 0 && zoomIndex < len(m.WeekEventGrids) {
		dayLayout := m.WeekEventGrids[zoomIndex].DayLayouts[day]
		if cellLayout, exists := dayLayout.CellLayouts[cellIndex]; exists {
			return cellLayout
		}
	}

	// Return empty layout if no events in this cell
	return types.CellLayout{}
}

// cursorInCell checks if the cursor is positioned at the specified cell
func cursorInCell(day int, cellIndex int, m types.WeekModel) bool {
	return day == m.Cursor.Day && cellIndex == m.Cursor.Cell
}

// isInVisualSelection checks if a given day/cell is within the visual selection range using time-based logic
func isInVisualSelection(day, cellIndex int, m types.WeekModel) bool {
	if !m.IsVisualMode {
		return false
	}

	// Convert current cell to time for comparison
	currentTime := util.CellToTime(cellIndex, m.CurrentZoom, m.CurrentlyViewedWeek, day)

	// Convert anchor to time (anchor might be in different week)
	anchorTime := util.CellToTime(m.VisualAnchor.Cell, m.CurrentZoom, m.VisualAnchorWeek, m.VisualAnchor.Day)

	// Convert cursor to time
	cursorTime := util.CellToTime(m.Cursor.Cell, m.CurrentZoom, m.CurrentlyViewedWeek, m.Cursor.Day)

	// Check if current time is between anchor and cursor (inclusive)
	minTime := anchorTime
	maxTime := cursorTime
	if cursorTime.Before(anchorTime) {
		minTime = cursorTime
		maxTime = anchorTime
	}

	return util.TimeIsWithinRange(currentTime, minTime, maxTime)
}

// RenderGridCell renders a single grid cell (empty or with events)
func RenderGridCell(m types.WeekModel, day, cellIndex int, content string, searchResults []*types.EventInstance) string {
	var style lipgloss.Style

	// Add events to cell
	cellLayout := GetCellLayout(m, day, cellIndex)

	// Determine background color and selection state
	isCursor := cursorInCell(day, cellIndex, m)
	isVisuallySelected := isInVisualSelection(day, cellIndex, m)
	bgColor := week_view_shared.DetermineCellBackgroundColor(isCursor, isVisuallySelected, m.CachedBackgroundSelectedColor)

	// Determine cell styling based on whether it has events
	if len(cellLayout) > 0 {
		style, content = renderCellWithEvents(m, cellLayout, isCursor, isVisuallySelected, bgColor, searchResults)
	} else {
		style = renderEmptyGridCell(m.CachedCellWidth, bgColor)
	}

	return style.Render(content)
}

// RenderDayColumn renders an entire day column containing multiple cells
func RenderDayColumn(m types.WeekModel, daysFromFirstDayOfWeek int, viewportStartCellIndex int, numRows int, searchResults []*types.EventInstance) string {
	var cells []string

	for i := range numRows {
		cellIndex := viewportStartCellIndex + i
		cell := RenderGridCell(m, daysFromFirstDayOfWeek, cellIndex, "", searchResults)
		cells = append(cells, cell)
	}
	return lipgloss.JoinVertical(lipgloss.Left, cells...)
}

// renderCellWithEvents renders a grid cell that contains events
func renderCellWithEvents(m types.WeekModel, cellLayout []types.PositionedEvent, isCursor, isVisuallySelected bool, bgColor string, searchResults []*types.EventInstance) (lipgloss.Style, string) {
	// Cell has events - don't apply background to cell, let individual events/columns handle it
	style := week_view_shared.CellStyle.Width(m.CachedCellWidth)

	// Always pass the cursor's event column - the rendering function will decide when to use it
	selectedEventColumn := -1
	if isCursor {
		selectedEventColumn = m.Cursor.EventColumn
	}

	// For empty columns, we always pass the cursor highlight color so it can be used when needed
	emptyColumnHighlightColor := m.CachedBackgroundSelectedColor

	searchActive := m.SearchActive || m.SearchLocked
	content := RenderEventsInCell(cellLayout, m.CachedContentWidth, selectedEventColumn, isVisuallySelected, bgColor, emptyColumnHighlightColor, isCursor, searchResults, searchActive, m.EventTextBlack, m.CachedBackgroundColor)

	return style, content
}

// renderEmptyGridCell renders a grid cell with no events
func renderEmptyGridCell(cellWidth int, bgColor string) lipgloss.Style {
	if bgColor != "" {
		return week_view_shared.CellStyle.Width(cellWidth).Background(lipgloss.Color(bgColor))
	}
	return week_view_shared.CellStyle.Width(cellWidth)
}