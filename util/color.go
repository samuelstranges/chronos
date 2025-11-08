// Package color provides color utilities and conversions for the application.
package util

import (
	"image/color"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
	types "github.com/samuelstranges/chronos/types"
)

// StandardizeHex ensures a hex color always has exactly one # prefix
// Handles cases: "FF0000", "#FF0000", "##FF0000" -> "#FF0000"
func StandardizeHex(hexColor string) string {
	if hexColor == "" {
		return ""
	}

	// Remove all # prefixes, then add exactly one
	cleaned := strings.TrimLeft(hexColor, "#")
	if cleaned == "" {
		return ""
	}

	return "#" + cleaned
}

// Color mappings for single letters (26 colors for A-Z)
// All colors styled to match Tokyo Night aesthetic: soft, muted, pleasant tones
var letterColors = map[string]string{
	"d": "#c9a87c",
	"a": "#e0af68",
	"y": "#e0af68",
	"q": "#e8c478",
	"o": "#ff9e64",
	"e": "#f5948e",
	"r": "#f7768e",
	"m": "#e898e8",
	"f": "#f5a4d0",
	"h": "#ffa8d0",
	"v": "#d8a8f7",
	"p": "#bb9af7",
	"u": "#a898f7",
	"i": "#b0a8f0",
	"z": "#9888cc",
	"l": "#c8b8f7",
	"j": "#a4d67c",
	"g": "#9ece6a",
	"t": "#6ed8dc",
	"c": "#5ed2ff",
	"n": "#7da4f7",
	"b": "#6ea3fe",
	"x": "#98a4b0",
	"k": "#8a8a8a",
	"s": "#d4d4d4",
	"w": "#cfc9c2",
}

func GetBackgroundColorHex() string {
	bgColor, err := lipgloss.BackgroundColor(os.Stdin, os.Stdout)
	if err != nil {
		// fallback - silently use default
		return types.ColorBlack
	}
	return ToHex(bgColor)
}

// ParseColor converts color input (hex, name, or letter) to hex value
// Returns empty string if input is invalid
func ParseColor(input string) string {
	if input == "" {
		return ""
	}

	input = strings.TrimSpace(input)

	// Check if it's a single letter first (our custom mapping)
	if len(input) == 1 {
		lowerInput := strings.ToLower(input)
		if hexColor, exists := letterColors[lowerInput]; exists {
			return hexColor
		}
		// Single letter not in our map is invalid
		return ""
	}

	// Try go-colorful for hex strings - it has proper error handling
	// go-colorful.Hex() expects color WITH # prefix
	standardInput := StandardizeHex(input)
	c, err := colorful.Hex(standardInput)
	if err != nil {
		return "" // Invalid input
	}

	return StandardizeHex(c.Hex())
}

// ToHex converts a color.Color to hex string using go-colorful
func ToHex(c color.Color) string {
	if c == nil {
		return types.ColorBlack
	}

	// Use go-colorful's MakeColor for direct conversion
	colorfulColor, ok := colorful.MakeColor(c)
	if !ok {
		return types.ColorBlack // Fallback for invalid colors
	}

	return StandardizeHex(colorfulColor.Hex())
}

// BrightenColor increases the brightness of a hex color by the given percentage
// For very dark colors or edge cases, this blends with white instead of pure brightening
func BrightenColor(hexColor string, percentage float64) string {
	// go-colorful.Hex() expects color WITH # prefix
	standardHex := StandardizeHex(hexColor)
	c, err := colorful.Hex(standardHex)
	if err != nil {
		return hexColor // Return original if invalid
	}

	// Create white color and blend with original
	white := colorful.Color{R: 1, G: 1, B: 1} // RGB white
	// Fixed 50% white blend for excellent visibility on all backgrounds
	blendAmount := 0.5 // 50% white blend
	blended := c.BlendRgb(white, blendAmount)

	return StandardizeHex(blended.Hex())
}

// ApplyBlueTint shifts a hex color toward blue hue while preserving brightness
func ApplyBlueTint(hexColor string, tintStrength float64) string {
	// go-colorful.Hex() expects color WITH # prefix
	standardHex := StandardizeHex(hexColor)
	c, err := colorful.Hex(standardHex)
	if err != nil {
		return hexColor // Return original if invalid
	}

	// Create blue color and blend
	blue := colorful.Color{R: 0, G: 0, B: 1}                     // RGB blue
	mixAmount := min(1.0, tintStrength*types.ColorDarkenPercent) // Clamp mixing to 30% max
	mixed := c.BlendRgb(blue, mixAmount)

	return StandardizeHex(mixed.Hex())
}
