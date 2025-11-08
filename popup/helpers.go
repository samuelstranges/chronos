package popup

import (
	"github.com/charmbracelet/lipgloss/v2"
	util "github.com/samuelstranges/chronos/util"
)

// CreateColorPreview creates a visual color preview block using foreground color and block characters
func CreateColorPreview(color string) string {
	if color == "" {
		return "  " // Always return 2 spaces for alignment
	}

	// Parse and validate the color
	hexColor := util.ParseColor(color)
	if hexColor == "" {
		return "  " // Return 2 spaces if invalid color
	}

	// Create a small color preview using foreground color with full block characters
	colorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(hexColor))

	return colorStyle.Render("██") // Two full block characters
}