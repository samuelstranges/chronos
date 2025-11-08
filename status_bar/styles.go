package status_bar

import "github.com/charmbracelet/lipgloss/v2"

var (
	// Mode indicator style (left corner) - changes color based on mode
	modeStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Bold(true)

	// Main status bar (dark background for middle section)
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#202328")).
			Foreground(lipgloss.Color("#bbc2cf")).
			PaddingLeft(1).
			PaddingRight(1)

	// Key sequence indicator (right side)
	keySequenceStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#51afef")).
				PaddingLeft(1).
				PaddingRight(1).
				Bold(true)

	// Error indicator style (right side, red background)
	errorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ec5f67")).
			PaddingLeft(1).
			PaddingRight(1).
			Bold(true)
)
