package keybinds_hints

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

const (
	hardcoded_description_len    = 18
	hardcoded_keybind_string_len = 1
)

// Simple styles
var (
	hintKeyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true)
	hintFolderKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	hintDescStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
)

func getKeyhintDescription(binding KeyBinding, width int) string {
	// Format description to exactly hardcoded_description_len characters
	desc := binding.Description
	if len(desc) > width {
		desc = desc[:width-len("...")] + "..."
	} else {
		desc = desc + strings.Repeat(" ", width-len(desc))
	}
	return desc
}

func getKeybindString(binding KeyBinding) string {
	keyChar := binding.Key

	// Handle special key symbols
	switch keyChar {
	case "tab":
		return "⇥"
	case "enter":
		return "↵"
	case "escape":
		return "⎋"
	case "left":
		return "←"
	case "right":
		return "→"
	case "up":
		return "↑"
	case "down":
		return "↓"
	default:
		// Take first character for regular keys
		if len(keyChar) > hardcoded_keybind_string_len {
			keyChar = keyChar[:hardcoded_keybind_string_len]
		}
		return keyChar
	}
}

// renderHint returns a string for a single keybinding - exactly hintWidth characters
func renderHint(binding KeyBinding) string {
	// Choose key style based on whether this is a folder
	var keyStyled string
	if binding.Action == "enter_layer" {
		keyStyled = hintFolderKeyStyle.Render(getKeybindString(binding))
	} else {
		keyStyled = hintKeyStyle.Render(getKeybindString(binding))
	}

	descStyled := hintDescStyle.Render(getKeyhintDescription(binding, hardcoded_description_len))
	return " " + keyStyled + " " + descStyled + " "
}

// renderRow takes keybindings and window width, returns row string and remaining keybindings
func renderRow(bindings []KeyBinding, windowWidth int) (string, []KeyBinding) {
	if len(bindings) == 0 {
		return "", nil
	}

	// Calculate how many hints can fit using lipgloss.Width for accurate display width
	maxHints := windowWidth / lipgloss.Width(renderHint(bindings[0])) // using len of first hint
	maxHints = min(maxHints, len(bindings))                           // cant show more hints than what exist
	maxHints = max(maxHints, 1)                                       // try to always show 1

	// Render only the hints that fit
	var rowContent string
	for i := 0; i < maxHints; i++ {
		rowContent = rowContent + renderHint(bindings[i])
	}

	// Return remaining bindings that didn't fit
	var remaining []KeyBinding
	if maxHints < len(bindings) {
		remaining = bindings[maxHints:]
	}

	return rowContent, remaining
}

// RenderKeyhints renders the complete keyhints display with header and rows
func RenderKeyhints(bindings []KeyBinding, windowWidth int, showAll bool, modeInfo types.ModeInfo, cachedBgColor string) string {
	if len(bindings) == 0 {
		return ""
	}

	var lines []string

	// First line: mode header
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(modeInfo.BackgroundColor)).
		Foreground(lipgloss.Color(modeInfo.ForegroundColor)).
		Bold(true).
		Width(windowWidth)

	headerLine := headerStyle.Render(modeInfo.Name)
	lines = append(lines, headerLine)

	// Content lines: render rows of keybindings (use cached background color)
	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(cachedBgColor)).
		Width(windowWidth)

	remainingBindings := bindings
	maxRows := 3 // hardcode max rows for limited view
	if showAll {
		maxRows = 999 // effectively unlimited
	}

	for row := 0; row < maxRows && len(remainingBindings) > 0; row++ {
		rowContent, remaining := renderRow(remainingBindings, windowWidth)
		contentLine := contentStyle.Render(rowContent)
		lines = append(lines, contentLine)
		remainingBindings = remaining
	}

	// Add "more" indicator if there are still remaining bindings and not showing all
	if len(remainingBindings) > 0 && !showAll {
		moreText := fmt.Sprintf("... and %d more (press space for all)", len(remainingBindings))
		// Truncate to fit window width (account for padding)
		moreText = util.TruncateContent(moreText, windowWidth-4)

		moreStyle := contentStyle.
			Foreground(lipgloss.Color("#fab387")).
			Italic(true)
		moreLine := moreStyle.Render(moreText)
		lines = append(lines, moreLine)
	}

	return strings.Join(lines, "\n")
}
