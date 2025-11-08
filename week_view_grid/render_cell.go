package week_view_grid

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/search_parser"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_shared"
)

// getEventTitle extracts title from ical event and adds recurring indicator
func getEventTitle(event *ical.Event) string {
	// WORKAROUND: Handle nil events from pre-allocated grid cells that weren't filled
	// This occurs because CellLayouts are created with make([]PositionedEvent, maxColumn+1)
	// which initializes unfilled positions with zero-value PositionedEvents (Event: nil)
	if event == nil {
		return ""
	}

	title := util.GetEventTitle(event)
	if util.IsRecurring(event) {
		title = "↻ " + title
	}

	return title
}

// getEventBackgroundRenderColor applies highlighting priority: search > visual > cursor
func getEventBackgroundRenderColor(event *ical.Event, calendarColor string, cursorOnEvent, isVisualSelection, searchActive bool, searchResults []*types.EventInstance) string {
	// Get event's color
	backgroundColor := util.GetEventColor(event, calendarColor)

	// Apply highlighting in priority order: search > visual selection, then brighten for cursor
	if searchActive {
		// Search is active - highlight matches and grey out non-matches
		if search_parser.IsEventInSearchResultsByInstance(event, searchResults) {
			// Search result match - red background
			backgroundColor = "#ff0000" // TODO: move to types
		} else {
			// Non-matching event - grey background
			backgroundColor = "#808080" // TODO: move to types
		}
	} else if isVisualSelection {
		// Apply visual selection blue tint
		backgroundColor = util.ApplyBlueTint(backgroundColor, types.CellOpacity)
	}

	// After applying search/visual colors, brighten for cursor selection
	if cursorOnEvent {
		backgroundColor = util.BrightenColor(backgroundColor, types.BrightenEventsPercent)
	}

	return backgroundColor
}

// renderEmptyColumn renders an empty event column with appropriate highlighting
func renderEmptyColumn(width int, isSelectedEvent, isCursorCell, isVisualSelection bool, emptyColumnHighlightColor, baseBgColor string) string {
	blankStyle := lipgloss.NewStyle().Width(width)

	var blankBgColor string
	// For empty columns, check if this is the cursor position AND the right column
	if isCursorCell && isSelectedEvent {
		blankBgColor = emptyColumnHighlightColor
	} else if isVisualSelection {
		blankBgColor = baseBgColor
	}

	if blankBgColor != "" {
		blankStyle = blankStyle.Background(lipgloss.Color(blankBgColor))
	}
	return blankStyle.Render("")
}

// renderEventColumn renders an actual event column with proper content and styling
func renderEventColumn(event types.PositionedEvent, eventTitle string, width int, isSelectedEvent, isVisualSelection, searchActive bool, searchResults []*types.EventInstance, eventTextBlack bool, cachedBgColor string) string {
	// Prepare content - use continuation indicator for non-start cells
	content := eventTitle
	if !event.IsStartCell {
		content = "┊" // continuation indicator
	}

	// Truncate content to fit the column width
	content = util.TruncateContent(content, width)

	// Determine background color with all highlighting logic
	var backgroundColor string
	if event.Instance != nil && event.Instance.OriginalEvent != nil {
		backgroundColor = getEventBackgroundRenderColor(
			event.Instance.OriginalEvent, event.CalendarColor, isSelectedEvent, isVisualSelection,
			searchActive, searchResults,
		)
	} else {
		backgroundColor = "#ff0000" // Red background for error case
	}

	// Apply styling
	eventStyle := lipgloss.NewStyle().Background(lipgloss.Color(backgroundColor)).Width(width)
	
	// Set text color based on toggle
	if eventTextBlack {
		eventStyle = eventStyle.Foreground(lipgloss.Color(cachedBgColor))
	}
	
	return eventStyle.Render(content)
}

func RenderEventsInCell(events []types.PositionedEvent, contentWidth int, selectedEventColumn int, isVisualSelection bool, baseBgColor string, emptyColumnHighlightColor string, isCursorCell bool, searchResults []*types.EventInstance, searchActive bool, eventTextBlack bool, cachedBgColor string) string {
	if len(events) == 0 {
		return ""
	}

	// Build each column cell (event or blank)
	columnStrings := make([]string, len(events))

	widthDist := week_view_shared.CalculateColumnWidths(contentWidth, len(events))

	for i, event := range events {
		width := widthDist.GetColumnWidth(i)

		// In visual mode at cursor position, ALL event columns should be brightened
		isSelectedEvent := selectedEventColumn == i || (isVisualSelection && isCursorCell)
		var eventTitle string
		if event.Instance != nil && event.Instance.OriginalEvent != nil {
			eventTitle = getEventTitle(event.Instance.OriginalEvent)
		} else {
			eventTitle = "" // Empty column - no event here
		}

		if eventTitle == "" {
			columnStrings[i] = renderEmptyColumn(width, isSelectedEvent, isCursorCell, isVisualSelection, emptyColumnHighlightColor, baseBgColor)
		} else {
			columnStrings[i] = renderEventColumn(event, eventTitle, width, isSelectedEvent, isVisualSelection, searchActive, searchResults, eventTextBlack, cachedBgColor)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, columnStrings...)
}
