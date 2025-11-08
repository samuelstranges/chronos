// Package statusbar provides status bar rendering functionality for the Chronos application.
package status_bar

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/keybinds"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/week_view_grid"
	"github.com/samuelstranges/chronos/search_parser"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view"
)

// RenderStatusBar renders a lualine-style status bar
func RenderStatusBar(
	weekModel *types.WeekModel, vimState *keybinds.VimState, width int, searchResults []*types.EventInstance,
) string {
	// Defensive checks for nil pointers
	if weekModel == nil || vimState == nil {
		return statusBarStyle.Width(width).Render("Error: nil pointer")
	}

	// Create content
	modeElement := createModeIndicator(vimState)
	eventInfo := createMiddleContent(weekModel, searchResults) // Middle: Search info (if active) or event info
	rightElement := createRightElement(weekModel, vimState)    // error or key sequence

	// Calculate widths for spacing
	modeWidth := lipgloss.Width(modeElement)
	rightWidth := lipgloss.Width(rightElement)
	middleWidth := week_view.CalculateRemainingWidth(width, modeWidth+rightWidth)

	// Create middle section with dark background
	middleElement := statusBarStyle.Width(middleWidth).Render(eventInfo)

	// Combine all elements
	finalElements := []string{modeElement, middleElement}
	if rightElement != "" {
		finalElements = append(finalElements, rightElement)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, finalElements...)
}

// createModeIndicator creates the left mode indicator element
func createModeIndicator(vimState *keybinds.VimState) string {
	currentMode := vimState.GetMode()
	modeInfo := currentMode.GetModeInfo()

	return modeStyle.
		Background(lipgloss.Color(modeInfo.BackgroundColor)).
		Foreground(lipgloss.Color(modeInfo.ForegroundColor)).
		Render(modeInfo.Name)
}

// createRightElement creates the right side element (error or key sequence)
func createRightElement(weekModel *types.WeekModel, vimState *keybinds.VimState) string {
	errMsgEmpty := weekModel.ErrorMessage != ""
	if errMsgEmpty && time.Now().Before(weekModel.ErrorExpiry) {
		return errorStyle.Render(weekModel.ErrorMessage)
	} else if keySeq := vimState.GetStatusText(); keySeq != "" {
		return keySequenceStyle.Render(keySeq)
	}
	return ""
}

// createMiddleContent creates the middle content (search info or event info)
func createMiddleContent(weekModel *types.WeekModel, searchResults []*types.EventInstance) string {
	if weekModel.SearchActive {
		return createSearchInfo(weekModel, searchResults)
	}
	return createEventInfo(weekModel)
}

// createSearchInfo creates search input and results count display
func createSearchInfo(weekModel *types.WeekModel, searchResults []*types.EventInstance) string {
	if weekModel.SearchInput != "" {
		_, phrase, valid := search_parser.ParseSearchInput(weekModel.SearchInput)
		if valid && phrase != "" {
			count := len(searchResults)
			return fmt.Sprintf("%s (%d matching event/s)", weekModel.SearchInput, count)
		} else {
			// Show the input even if not valid yet
			return weekModel.SearchInput
		}
	}
	return ""
}

// createEventInfo creates current event display
func createEventInfo(weekModel *types.WeekModel) string {
	if event := week_view_grid.GetEventUnderCursor(weekModel); event != nil && event.Instance != nil && event.Instance.OriginalEvent != nil {
		// Use the event to get title safely
		eventInfo := util.GetEventTitle(event.Instance.OriginalEvent)
		// Get time range using timezone-safe methods
		startTime, endTime, err := timezone.GetEventTimes(event.Instance.OriginalEvent)
		if err == nil {
			timeRange := fmt.Sprintf("%s - %s", startTime.Format("15:04"), endTime.Format("15:04"))
			eventInfo = fmt.Sprintf("%s (%s)", eventInfo, timeRange)
		}
		return eventInfo
	}
	return ""
}
