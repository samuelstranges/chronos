package week_view_shared

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

var (
	// Default cell style
	CellStyle = lipgloss.NewStyle().
			Height(1).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			PaddingLeft(types.CellPaddingLeft).
			PaddingRight(types.CellPaddingRight)

	TimeStyle = CellStyle.
			Align(lipgloss.Right).
			PaddingRight(types.TimeColumnRightPadding).
			PaddingLeft(types.TimeColumnLeftPadding).
			Width(types.TimeColumnAndPaddingAndBorderWidth)

	TimeStyleCurrentTimeCell = TimeStyle.
					Foreground(lipgloss.Color(types.CurrentTimeRowColor))

	TimeStyleCursorCell = TimeStyle.
				Foreground(lipgloss.Color(types.TimeCellCursorRowColor))

	GenericHeaderStyle = CellStyle.
				Height(1).
				Bold(true).
				BorderBackground(lipgloss.Color(types.HeaderBackgroundColor)).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color(types.HeaderBackgroundColor))

	TimeStyleHeader = TimeStyle.
			BorderBackground(lipgloss.Color(types.HeaderBackgroundColor)).
			Background(lipgloss.Color(types.HeaderBackgroundColor))

	HeaderRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(types.HeaderBackgroundColor))

	MonthYearHeaderStyle = lipgloss.NewStyle().
				Height(1).
				Bold(true).
				Align(lipgloss.Center).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("235"))
)

// WidthDistribution represents how to distribute width across columns
type WidthDistribution struct {
	BaseWidth      int
	RemainderChars int
	TotalColumns   int
}

// CalculateColumnWidths determines how to distribute width across columns
// Returns base width for each column and number of columns that get an extra character
func CalculateColumnWidths(contentWidth, numColumns int) WidthDistribution {
	if numColumns <= 0 {
		return WidthDistribution{BaseWidth: contentWidth, RemainderChars: 0, TotalColumns: 0}
	}

	baseWidth := contentWidth / numColumns
	remainderChars := contentWidth % numColumns

	return WidthDistribution{
		BaseWidth:      baseWidth,
		RemainderChars: remainderChars,
		TotalColumns:   numColumns,
	}
}

// GetColumnWidth returns the width for a specific column index
func (wd WidthDistribution) GetColumnWidth(columnIndex int) int {
	if columnIndex < wd.RemainderChars {
		// Give extra character to first remainderChars columns
		return wd.BaseWidth + 1
	}
	return wd.BaseWidth
}

// CalculateCellWidth calculates cell and content widths for the week view
func CalculateCellWidth(modelWidth int) (cellWidth, contentWidth int) {
	availableWidth := modelWidth - types.TimeColumnAndPaddingAndBorderWidth

	// In lipgloss v2, Width() sets the total element width (including borders)
	// So we just divide the available width by number of columns
	cellWidth = availableWidth / types.DaysPerWeek
	contentWidth = cellWidth - CellStyle.GetHorizontalFrameSize() // Subtract border and padding for content

	return cellWidth, contentWidth
}

// DetermineCellBackgroundColor calculates background color for cursor/visual selection
func DetermineCellBackgroundColor(isCursor, isVisuallySelected bool, baseSelectedColor string) string {
	if isCursor {
		// Cursor highlighting takes precedence - bright highlighting
		return baseSelectedColor
	} else if isVisuallySelected {
		// Visual selection - strong blue tint to background for clear distinction
		return util.ApplyBlueTint(baseSelectedColor, types.ReducedOpacity)
	}
	return ""
}
