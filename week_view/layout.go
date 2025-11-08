// Package week handles layout calculations and distribution for week view components.
package week_view

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

const (
	centeringDivisor = 2
)


// CenterContent calculates positioning to center content within available space
func CenterContent(contentWidth, contentHeight, availableWidth, availableHeight int) (x, y int) {
	x = (availableWidth - contentWidth) / centeringDivisor
	y = (availableHeight - contentHeight) / centeringDivisor

	return max(0, x), max(0, y) // Ensure non-negative positioning
}


// DistributeHorizontalSpace creates spacing between elements
func DistributeHorizontalSpace(elements []string, totalWidth int, minSpacing int) string {
	if len(elements) == 0 {
		return ""
	}
	if len(elements) == 1 {
		return elements[0]
	}

	// Calculate total content width
	totalContentWidth := 0
	for _, element := range elements {
		totalContentWidth += lipgloss.Width(element)
	}

	// Calculate available space for gaps
	availableSpace := totalWidth - totalContentWidth
	numGaps := len(elements) - 1

	// Determine spacing per gap
	var spacingPerGap int
	if numGaps > 0 && availableSpace > 0 {
		spacingPerGap = max(availableSpace/numGaps, minSpacing)
	} else {
		spacingPerGap = minSpacing
	}

	// Join elements with calculated spacing
	spacer := strings.Repeat(" ", spacingPerGap)
	result := strings.Join(elements, spacer)

	return result
}


// CalculateRemainingWidth calculates remaining width after accounting for used width
func CalculateRemainingWidth(totalWidth, usedWidth int) int {
	remaining := totalWidth - usedWidth
	return max(0, remaining)
}
