// Package commands provides vim-style key sequence handling for Chronos
// Inspired by https://github.com/kujtimiihoxha/vimtea (MIT License)
package keybinds

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/types"
)

// KeyhintAdapter provides an adapter for keyhints to access vim registry
type KeyhintAdapter struct {
	registry *CommandRegistry
}

// NewKeyhintAdapter creates a new adapter for keyhints
func NewKeyhintAdapter(registry *CommandRegistry) *KeyhintAdapter {
	return &KeyhintAdapter{registry: registry}
}

// GetKeyhintBindingsForMode returns bindings for keyhints (adapter method)
func (a *KeyhintAdapter) GetKeyhintBindingsForMode(mode types.VimMode) []types.VimCommandBinding {
	bindings := a.registry.GetBindingsForMode(mode)
	var result []types.VimCommandBinding

	for _, binding := range bindings {
		keyhintBinding := types.VimCommandBinding{
			Key:         binding.Key,
			Description: binding.Description,
			Modes:       binding.Modes,
		}
		result = append(result, keyhintBinding)
	}

	return result
}

// GetKeyhintFolders returns folders for keyhints (adapter method)
func (a *KeyhintAdapter) GetKeyhintFolders() map[string]types.VimCommandFolder {
	folders := a.registry.folders
	result := make(map[string]types.VimCommandFolder)

	// Already in deterministic order since we use slices
	for _, folder := range folders {
		keyhintFolder := types.VimCommandFolder{
			Prefix:      folder.Prefix,
			Description: folder.Description,
		}
		result[folder.Prefix] = keyhintFolder
	}

	return result
}

// HandleKey processes a key input and returns command if complete
// Now handles ALL keys including UI keys (space, escape, hints) to centralize logic
func (v *VimState) HandleKey(msg tea.KeyMsg, registry *CommandRegistry, isFormVisible bool) types.VimActionMsg {
	keyStr := msg.String()

	// If form is visible, only handle escape to close form
	if isFormVisible {
		if keyStr == "esc" {
			return types.VimActionMsg{Action: "close_form"}
		}
		return types.VimActionMsg{} // Let form handle all other keys
	}

	// Handle space key for key hints
	if keyStr == "space" {
		return types.VimActionMsg{Action: "toggle_hints"}
	}

	// Handle escape - reset vim state, exit visual mode, or clear search
	if keyStr == "esc" {
		v.reset()

		// If in visual mode, return to normal mode
		if v.Mode == types.ModeVisual {
			return types.VimActionMsg{Action: "exit_visual"}
		}

		// If in normal mode, hide hints (search clearing handled elsewhere)
		if v.Mode == types.ModeNormal {
			return types.VimActionMsg{Action: "hide_hints"}
		}

		// If in search mode, exit search
		if v.Mode == types.ModeSearch {
			return types.VimActionMsg{Action: "exit_search_mode"}
		}

		return types.VimActionMsg{}
	}

	// Regular vim command handling
	if actionMsg := v.handleVimCommand(msg, registry); actionMsg.Action != "" {
		v.reset()
		return actionMsg
	}

	// Check if vim is building a sequence (update hints immediately)
	if v.GetStatusText() != "" {
		return types.VimActionMsg{Action: "update_hints_for_sequence"}
	}

	// Key not handled
	return types.VimActionMsg{}
}

// handleVimCommand processes vim-specific commands (extracted from original HandleKey logic)
func (v *VimState) handleVimCommand(msg tea.KeyMsg, registry *CommandRegistry) types.VimActionMsg {
	keyStr := msg.String()

	// Handle numeric prefixes
	if len(v.KeySequence) == 0 && keyStr >= "1" && keyStr <= "9" {
		count, _ := strconv.Atoi(keyStr)
		v.CountPrefix = count
		v.KeySequence = append(v.KeySequence, keyStr)
		v.updateStatusText()
		return types.VimActionMsg{}
	} else if len(v.KeySequence) > 0 && keyStr >= "0" && keyStr <= "9" {
		// Check if continuing numeric prefix
		allDigits := true
		for _, k := range v.KeySequence {
			if k < "0" || k > "9" {
				allDigits = false
				break
			}
		}

		if allDigits {
			countStr := strings.Join(v.KeySequence, "") + keyStr
			count, _ := strconv.Atoi(countStr)
			v.CountPrefix = count
			v.KeySequence = append(v.KeySequence, keyStr)
			v.updateStatusText()
			return types.VimActionMsg{}
		}
	}

	// Add key to sequence
	v.KeySequence = append(v.KeySequence, keyStr)
	seq := strings.Join(v.KeySequence, "")

	// Check for exact match in registry (includes g-commands now)
	if actionMsg := registry.FindExact(seq, v.Mode, v.CountPrefix); actionMsg.Action != "" {
		return actionMsg
	}

	// Check if prefix - wait for more input
	if registry.IsPrefix(seq, v.Mode) {
		v.updateStatusText()
		return types.VimActionMsg{}
	}

	// No match found - reset
	v.reset()
	return types.VimActionMsg{}
}
