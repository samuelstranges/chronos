package keybinds_hints

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/types"
)

// KeyHintTimeoutMsg is sent after a delay to show key hints
type KeyHintTimeoutMsg struct{}

// CreateKeyHintTimeout returns a command that sends KeyHintTimeoutMsg after KeyHintTimeoutSeconds
func CreateKeyHintTimeout() tea.Cmd {
	return tea.Tick(time.Duration(types.KeyHintTimeoutSeconds)*time.Second, func(time.Time) tea.Msg {
		return KeyHintTimeoutMsg{}
	})
}

// HandleKeyHintTimeout shows key hints after timeout if still building a sequence
func HandleKeyHintTimeout(keyHints *KeyHintSystem, vimMode types.VimMode, statusText string, waitingForTimeout *bool) tea.Cmd {
	if statusText != "" {
		keyHints.SetMode(ConvertVimMode(int(vimMode)))
		keyHints.ShowForVimState(statusText)
	}
	*waitingForTimeout = false
	return nil
}

// HandleNavigateHints tries to navigate to a layer with the given key
func HandleNavigateHints(keyHints *KeyHintSystem, action string) bool {
	if strings.HasPrefix(action, "navigate_hints:") {
		keyStr := strings.TrimPrefix(action, "navigate_hints:")
		return keyHints.NavigateToLayer(keyStr)
	}
	return false
}

// HandleNavigateHintsBack navigates back up in the keyhints hierarchy
func HandleNavigateHintsBack(keyHints *KeyHintSystem) tea.Cmd {
	if keyHints.IsVisible() {
		if keyHints.NavigateBack() {
			// Successfully navigated back, stay visible
			return nil
		} else {
			// At root level, hide hints
			keyHints.Hide()
			return nil
		}
	}
	return nil
}
