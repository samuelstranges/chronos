package keybinds_hints

import (
	"strings"

	"github.com/samuelstranges/chronos/types"
)

// VimMode imported from types package

// KeyBinding represents a complete key binding definition
type KeyBinding struct {
	Key         string          // The key sequence
	Description string          // Human readable description
	Action      string          // The action this binding triggers
	Example     string          // Example usage (optional)
	Modes       []types.VimMode // Which modes this binding is available in
}

// KeyLayer represents a layer in the key hierarchy
type KeyLayer struct {
	Name        string               // Layer name (e.g., "root", "g-commands")
	Prefix      string               // Key prefix that activates this layer
	Parent      *KeyLayer            // Parent layer (nil for root)
	Children    map[string]*KeyLayer // Child layers keyed by their trigger
	Bindings    []KeyBinding         // Direct bindings at this layer
	Description string               // Description of what this layer contains
}

// VimRegistry interface to access vim command registry
type VimRegistry interface {
	GetKeyhintBindingsForMode(mode types.VimMode) []types.VimCommandBinding
	GetKeyhintFolders() map[string]types.VimCommandFolder
}

// KeyHintSystem manages the entire key hint hierarchy
type KeyHintSystem struct {
	currentMode types.VimMode
	visible     bool
	showAll     bool        // Whether to show all commands or just first few rows
	keySequence []string    // Current key sequence being built
	vimRegistry VimRegistry // Reference to vim command registry
}

// NewKeyHintSystem creates a new key hint system with vim registry
func NewKeyHintSystem(vimRegistry VimRegistry) *KeyHintSystem {
	system := &KeyHintSystem{
		currentMode: types.ModeNormal,
		visible:     false,
		keySequence: make([]string, 0),
		vimRegistry: vimRegistry,
	}

	// All bindings built dynamically from vim registry

	return system
}

// GetBindingsForMode returns all bindings available for the given mode at current layer
func (s *KeyHintSystem) GetBindingsForMode(mode types.VimMode) []KeyBinding {
	return s.buildDynamicBindingsForMode(mode)
}

// buildDynamicBindingsForMode builds bindings dynamically from vim registry
func (s *KeyHintSystem) buildDynamicBindingsForMode(mode types.VimMode) []KeyBinding {
	var result []KeyBinding

	// Get current layer context
	currentPrefix := s.GetKeySequence()

	if currentPrefix == "" {
		// Root level - show all direct commands and folders
		vimBindings := s.vimRegistry.GetKeyhintBindingsForMode(mode)
		folders := s.vimRegistry.GetKeyhintFolders()

		// Add folder entries (already in deterministic order from vim registry)
		for prefix, folder := range folders {
			folderBinding := KeyBinding{
				Key:         prefix,
				Description: folder.Description,
				Action:      "enter_layer",
				Modes:       []types.VimMode{mode},
			}
			result = append(result, folderBinding)
		}

		// Add direct commands (not starting with folder prefixes)
		for _, vimBinding := range vimBindings {
			if !s.startsWithFolder(vimBinding.Key) && !strings.HasPrefix(vimBinding.Key, "ctrl+") {
				keyBinding := KeyBinding{
					Key:         vimBinding.Key,
					Description: vimBinding.Description,
					Action:      "vim_command",
					Modes:       vimBinding.Modes,
				}
				result = append(result, keyBinding)
			}
		}

	} else {
		// Inside a folder - show commands with this prefix
		vimBindings := s.vimRegistry.GetKeyhintBindingsForMode(mode)
		for _, vimBinding := range vimBindings {
			if strings.HasPrefix(vimBinding.Key, currentPrefix) && len(vimBinding.Key) > len(currentPrefix) && !strings.HasPrefix(vimBinding.Key, "ctrl+") {
				// Remove prefix to get sub-key
				subKey := vimBinding.Key[len(currentPrefix):]

				// Strip pattern symbols (# characters) to show just the command key
				displayKey := strings.ReplaceAll(subKey, "#", "")

				keyBinding := KeyBinding{
					Key:         displayKey,
					Description: vimBinding.Description,
					Action:      "vim_command",
					Modes:       vimBinding.Modes,
				}
				result = append(result, keyBinding)
			}
		}
	}

	return result
}

// startsWithFolder checks if a key starts with any folder prefix
func (s *KeyHintSystem) startsWithFolder(key string) bool {
	folders := s.vimRegistry.GetKeyhintFolders()
	for prefix := range folders {
		if strings.HasPrefix(key, prefix) && len(key) > len(prefix) {
			return true
		}
	}
	return false
}

// NavigateToLayer navigates to a child layer (dynamic system)
func (s *KeyHintSystem) NavigateToLayer(key string) bool {
	// Check if this key is a valid folder
	folders := s.vimRegistry.GetKeyhintFolders()
	currentPrefix := s.GetKeySequence()
	targetPrefix := currentPrefix + key

	// Check if target prefix exists as a folder
	if _, exists := folders[targetPrefix]; exists {
		s.keySequence = append(s.keySequence, key)
		return true
	}

	return false
}

// NavigateBack goes back to parent layer (dynamic system)
func (s *KeyHintSystem) NavigateBack() bool {
	if len(s.keySequence) > 0 {
		s.keySequence = s.keySequence[:len(s.keySequence)-1]
		return true
	}
	return false
}

// ResetToRoot resets to the root layer (dynamic system)
func (s *KeyHintSystem) ResetToRoot() {
	s.keySequence = make([]string, 0)
	s.showAll = false // Reset to limited view
}

// Show displays the hint popup for current layer and mode
func (s *KeyHintSystem) Show() {
	s.visible = true
}

// ShowForVimState shows keyhints for the appropriate layer based on vim state
func (s *KeyHintSystem) ShowForVimState(vimStatusText string) {
	s.ResetToRoot()

	// Dynamically navigate to folder if it exists
	folders := s.vimRegistry.GetKeyhintFolders()
	if _, exists := folders[vimStatusText]; exists {
		s.NavigateToLayer(vimStatusText)
	}

	s.Show()
}

// Hide hides the hint popup
func (s *KeyHintSystem) Hide() {
	s.visible = false
}

// IsVisible returns whether the popup is currently visible
func (s *KeyHintSystem) IsVisible() bool {
	return s.visible
}

// IsShowingAll returns whether all commands are currently shown
func (s *KeyHintSystem) IsShowingAll() bool {
	return s.showAll
}

// SetMode changes the current vim mode
func (s *KeyHintSystem) SetMode(mode types.VimMode) {
	s.currentMode = mode
}

// ToggleShowAll toggles between showing limited and all commands
func (s *KeyHintSystem) ToggleShowAll() {
	s.showAll = !s.showAll
}

// GetCurrentLayer returns the current layer name (dynamic system)
func (s *KeyHintSystem) GetCurrentLayer() string {
	if len(s.keySequence) == 0 {
		return "root"
	}
	// Return the name of the current folder context
	folders := s.vimRegistry.GetKeyhintFolders()
	prefix := strings.Join(s.keySequence, "")
	if folder, exists := folders[prefix]; exists {
		return folder.Description
	}
	return prefix
}

// GetKeySequence returns the current key sequence path
func (s *KeyHintSystem) GetKeySequence() string {
	return strings.Join(s.keySequence, "")
}

// ConvertVimMode from vim package types (if needed for integration)
func ConvertVimMode(mode int) types.VimMode {
	switch mode {
	case 0: // commands.ModeNormal
		return types.ModeNormal
	case 1: // commands.ModeVisual
		return types.ModeVisual
	default:
		return types.ModeNormal
	}
}
