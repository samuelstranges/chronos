package util

import (
	"github.com/charmbracelet/lipgloss/v2"
)

// TruncateContent truncates content to fit within the specified maximum width
func TruncateContent(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(content) <= maxWidth {
		return content
	}
	return lipgloss.NewStyle().MaxWidth(maxWidth).Render(content)
}