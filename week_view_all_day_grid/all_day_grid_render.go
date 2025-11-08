package week_view_all_day_grid

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_shared"
)

const EventContinuedStr = "..."

func RenderAllDayEventSection(m types.WeekModel, counts []int) string {
	var rows []string

	rows = append(rows, renderAllDayCountRow(m, counts)) // header

	for rowIndex := range types.AllDayCells { // rows
		rows = append(rows, renderAllDayGridRow(m, m.AllDayEventGrid, rowIndex))
	}

	rows = append(rows, renderAllDayGridSeperator(m)) // separator

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderAllDayGridRow renders a single row of the all-day event grid using right borders
func renderAllDayGridRow(m types.WeekModel, allDayGrid types.AllDayEventGrid, rowIndex int) string {
	var elements []string

	// Add time column (empty for all-day events)
	elements = append(elements, week_view_shared.TimeStyle.Render(""))

	// Render first 6 days with border logic
	for day := range types.DaysPerWeek - 1 {
		currentEvent := getEventInCell(allDayGrid, rowIndex, day)
		nextEvent := getEventInCell(allDayGrid, rowIndex, day+1)
		borderColor, needsCustomBorder := getBorderColor(currentEvent, nextEvent)

		cell := renderAllDayGridCellWithBorder(allDayGrid, rowIndex, day, m.CachedCellWidth, m.CachedContentWidth, borderColor, needsCustomBorder, m)
		elements = append(elements, cell)
	}

	// Render last day (no right border logic needed)
	lastDay := types.DaysPerWeek - 1
	cell := renderAllDayGridCellWithBorder(allDayGrid, rowIndex, lastDay, m.CachedCellWidth, m.CachedContentWidth, "", false, m)
	elements = append(elements, cell)

	return lipgloss.JoinHorizontal(lipgloss.Top, elements...)
}

// renderAllDayGridCellWithBorder renders a single cell in the all-day grid with right border
func renderAllDayGridCellWithBorder(allDayGrid types.AllDayEventGrid, rowIndex, day, cellWidth, contentWidth int, borderColor string, needsCustomBorder bool, m types.WeekModel) string {
	// Check if this row has events
	if rowIndex >= len(allDayGrid.EventRows) {
		return week_view_shared.CellStyle.Width(cellWidth).Render("") // out of bounds?
	}

	// Find an event that spans this day
	eventsInRow := allDayGrid.EventRows[rowIndex]
	for _, event := range eventsInRow {
		if isDayInAllDayEventSpan(day, event) {
			return renderAllDayEventCellWithBorder(event, day, cellWidth, contentWidth, borderColor, needsCustomBorder, m.EventTextBlack, m.CachedBackgroundColor)
		}
	}

	return week_view_shared.CellStyle.Width(cellWidth).Render("") // no events
}

// isDayInAllDayEventSpan checks if a day falls within an all-day event's span
func isDayInAllDayEventSpan(day int, event types.AllDaySpanningEvent) bool {
	return day >= event.StartDay && day <= event.EndDay
}

// renderAllDayEventCellWithBorder renders a single cell of an all-day spanning event with right border
func renderAllDayEventCellWithBorder(event types.AllDaySpanningEvent, currentDay, cellWidth, contentWidth int, borderColor string, needsCustomBorder bool, eventTextBlack bool, cachedBgColor string) string {
	// Setup title
	title := util.GetEventTitle(event.Event)
	if currentDay != event.StartDay {
		title = EventContinuedStr // show ... to indicate continue
	}
	title = util.TruncateContent(title, contentWidth)

	// Setup style
	style := week_view_shared.CellStyle.Width(cellWidth).Background(lipgloss.Color(event.CalendarColor))
	if needsCustomBorder { // continues w/ next event
		style = style.BorderBackground(lipgloss.Color(borderColor))
	}
	
	// Set text color based on toggle
	if eventTextBlack {
		style = style.Foreground(lipgloss.Color(cachedBgColor))
	}

	return style.Render(title)
}

func getBorderColor(leftEvent, rightEvent *types.AllDaySpanningEvent) (string, bool) {
	if leftEvent != nil && rightEvent != nil && leftEvent.Event == rightEvent.Event {
		// Same event spans both cells
		return leftEvent.CalendarColor, true
	} else {
		return "", false
	}
}

// getEventInCell returns the event in a specific cell, or nil if no event
func getEventInCell(allDayGrid types.AllDayEventGrid, rowIndex, day int) *types.AllDaySpanningEvent {
	if rowIndex >= len(allDayGrid.EventRows) {
		return nil
	}

	// Find an event that spans this day
	for _, event := range allDayGrid.EventRows[rowIndex] {
		if isDayInAllDayEventSpan(day, event) {
			return &event
		}
	}

	return nil
}

// renderAllDayCountRow renders a header row showing the count of all-day events per day
func renderAllDayCountRow(m types.WeekModel, counts []int) string {
	var cells []string
	var content string

	// Add time column (empty)
	timeCell := week_view_shared.TimeStyleHeader.Render("")
	cells = append(cells, timeCell)

	// Count events for each day
	for day := range types.DaysPerWeek {
		if counts[day] == 0 {
			content = ""
		} else {
			content = fmt.Sprintf("%d all-day events", counts[day])
			content = util.TruncateContent(content, m.CachedContentWidth)
		}

		countCell := week_view_shared.GenericHeaderStyle.Width(m.CachedCellWidth).Render(content)
		cells = append(cells, countCell)
	}

	// Wrap the entire row with full-width header background
	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return week_view_shared.HeaderRowStyle.Width(m.Width).Render(row)
}

// renderAllDayGridSeperator renders a horizontal bar separator with border intersections
func renderAllDayGridSeperator(m types.WeekModel) string {
	timeString := strings.Repeat(" ", types.TimeColumnAndPaddingWidth) + "├"
	normalCell := strings.Repeat("─", m.CachedContentWidth) + "┼"
	rightCell := strings.Repeat("─", m.CachedContentWidth) + "┤"
	totalRowStr := timeString + strings.Repeat(normalCell, types.MaxDayIndex) + rightCell

	return lipgloss.NewStyle().Render(totalRowStr)
}
