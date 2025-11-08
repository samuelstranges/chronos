package keybinds

import (
	"slices"
	"strings"

	"github.com/samuelstranges/chronos/types"
)

// registerAllKeybinds registers all vim keybindings in a centralized location
func (r *CommandRegistry) registerAllKeybinds() {
	// Command folders/groups
	r.addFolder("g", "Go commands")
	r.addFolder("c", "Change commands")

	// Movement commands
	r.addKeybind(types.ModeNormal, "h", "move_left", "Move left")
	r.addKeybind(types.ModeNormal, "j", "move_down", "Move down")
	r.addKeybind(types.ModeNormal, "k", "move_up", "Move up")
	r.addKeybind(types.ModeNormal, "l", "move_right", "Move right")
	r.addKeybind(types.ModeNormal, "left", "move_left", "Move left")
	r.addKeybind(types.ModeNormal, "down", "move_down", "Move down")
	r.addKeybind(types.ModeNormal, "up", "move_up", "Move up")
	r.addKeybind(types.ModeNormal, "right", "move_right", "Move right")
	r.addKeybind(types.ModeNormal, "H", "navigate_to_previous_week", "Previous week")
	r.addKeybind(types.ModeNormal, "L", "navigate_to_next_week", "Next week")
	r.addKeybind(types.ModeNormal, "w", "move_to_next_event", "Next event")
	r.addKeybind(types.ModeNormal, "b", "move_to_previous_event", "Previous event")
	r.addKeybind(types.ModeNormal, "e", "move_to_end_of_event", "End of event")
	r.addKeybind(types.ModeNormal, "t", "jump_to_current_time", "Jump to now")
	r.addKeybind(types.ModeNormal, "G", "jump_to_end_of_day", "Go to end")
	r.addKeybind(types.ModeNormal, "gg", "jump_to_start_of_day", "Go to start")
	r.addKeybind(types.ModeNormal, "^", "jump_to_first_event_of_day", "Day's 1st event")
	r.addKeybind(types.ModeNormal, "$", "jump_to_last_event_of_day", "Day's last event")

	// Event management
	r.addKeybind(types.ModeNormal, "enter", "view_event_details", "Event details")
	r.addKeybind(types.ModeNormal, "o", "open_url", "Open event URL")

	// Application control
	r.addKeybind(types.ModeNormal, "q", "quit", "Quit")
	r.addKeybind(types.ModeNormal, "ctrl+c", "quit", "Quit")

	// Mode transitions
	r.addKeybind(types.ModeNormal, "v", "enter_visual", "Enter visual")
	r.addKeybind(types.ModeNormal, "/", "enter_search_mode", "Enter search")

	// Actions (normal mode)
	r.addKeybind(types.ModeNormal, "x", "cut", "Cut event")
	r.addKeybind(types.ModeNormal, "y", "copy", "Copy event")
	r.addKeybind(types.ModeNormal, "p", "paste", "Paste event")
	r.addKeybind(types.ModeNormal, "A", "toggle_agenda", "View agenda")
	r.addKeybind(types.ModeNormal, "tab", "toggle_all_day_grid", "All-Day Toggle")
	r.addKeybind(types.ModeNormal, "u", "undo", "Undo")
	r.addKeybind(types.ModeNormal, "U", "redo", "Redo")
	r.addKeybind(types.ModeNormal, "+", "zoom_in", "Zoom in")
	r.addKeybind(types.ModeNormal, "-", "zoom_out", "Zoom out")
	r.addKeybind(types.ModeNormal, "T", "toggle_event_text_color", "Toggle text color")

	// G-commands (navigation) - Pattern commands
	r.addKeybind(types.ModeNormal, "gh##", "go_to_hour", "Go to hour")
	r.addKeybind(types.ModeNormal, "gm##", "go_to_minute", "Go to minute")
	r.addKeybind(types.ModeNormal, "gt####", "go_to_time", "Go to time")
	r.addKeybind(types.ModeNormal, "gd##", "go_to_day", "Go to day")
	r.addKeybind(types.ModeNormal, "gM##", "go_to_month", "Go to month")
	r.addKeybind(types.ModeNormal, "gY####", "go_to_year", "Go to year")

	// a-command (add events) - unified dynamic form
	r.addKeybind(types.ModeNormal, "a", "add_event", "Add event")

	// C-commands (calendar)
	r.addKeybind(types.ModeNormal, "C", "calendar_popup", "Calendar popup")

	// c-commands (change event properties) using universal form system
	r.addKeybind(types.ModeNormal, "cc", "edit_event", "Change event")
	r.addKeybind(types.ModeNormal, "cC", "change_color", "Change color")
	r.addKeybind(types.ModeNormal, "cd", "change_duration", "Change duration")
	r.addKeybind(types.ModeNormal, "cD", "change_description", "Change desc")
	r.addKeybind(types.ModeNormal, "cs", "change_start", "Change start")
	r.addKeybind(types.ModeNormal, "ce", "change_end", "Change end")
	r.addKeybind(types.ModeNormal, "ct", "change_title", "Change title")
	r.addKeybind(types.ModeNormal, "cl", "change_location", "Change location")
	r.addKeybind(types.ModeNormal, "cL", "change_link", "Change link")

	// Visual mode commands
	r.addKeybind(types.ModeVisual, "h", "visual_move_left", "Move left")
	r.addKeybind(types.ModeVisual, "j", "visual_move_down", "Move down")
	r.addKeybind(types.ModeVisual, "k", "visual_move_up", "Move up")
	r.addKeybind(types.ModeVisual, "l", "visual_move_right", "Move right")
	r.addKeybind(types.ModeVisual, "left", "visual_move_left", "Move left")
	r.addKeybind(types.ModeVisual, "down", "visual_move_down", "Move down")
	r.addKeybind(types.ModeVisual, "up", "visual_move_up", "Move up")
	r.addKeybind(types.ModeVisual, "right", "visual_move_right", "Move right")
	r.addKeybind(types.ModeVisual, "H", "navigate_to_previous_week", "Previous week")
	r.addKeybind(types.ModeVisual, "L", "navigate_to_next_week", "Next week")
	r.addKeybind(types.ModeVisual, "w", "move_to_next_event", "Next event")
	r.addKeybind(types.ModeVisual, "b", "move_to_previous_event", "Previous event")
	r.addKeybind(types.ModeVisual, "e", "move_to_end_of_event", "End of event")
	r.addKeybind(types.ModeVisual, "G", "jump_to_end_of_day", "Go to end")
	r.addKeybind(types.ModeVisual, "gg", "jump_to_start_of_day", "Go to start")
	r.addKeybind(types.ModeVisual, "v", "exit_visual", "Exit visual")
	r.addKeybind(types.ModeVisual, "y", "copy_selection", "Copy selection")
	r.addKeybind(types.ModeVisual, "x", "cut_selection", "Cut selection")
	r.addKeybind(types.ModeVisual, "^", "jump_to_first_event_of_day", "First event of day")
	r.addKeybind(types.ModeVisual, "$", "jump_to_last_event_of_day", "Last event of day")
	r.addKeybind(types.ModeVisual, "o", "swap_selection_ends", "Swap ends")

	// Search mode commands
	r.addKeybind(types.ModeSearch, "esc", "exit_search_mode", "Exit search")
	r.addKeybind(types.ModeSearch, "ctrl+c", "exit_search_mode", "Exit search")
	r.addKeybind(types.ModeSearch, "enter", "execute_search", "Execute search")
	r.addKeybind(types.ModeSearch, "n", "search_next", "Next match")
	r.addKeybind(types.ModeSearch, "N", "search_previous", "Prev match")
}

// AddKeybind registers an action string for the specified mode
func (r *CommandRegistry) addKeybind(mode types.VimMode, key string, action string, description string) {
	// Check if this is a pattern command (contains # placeholders)
	if strings.Contains(key, "#") {
		// This is a pattern - add to pattern entries
		r.patternEntries = append(r.patternEntries, PatternEntry{
			Pattern: key,
			Action:  action,
		})

		// Register prefixes for the pattern (without the # parts)
		basePattern := strings.ReplaceAll(key, "#", "")
		for i := 1; i < len(basePattern); i++ {
			prefix := basePattern[:i]
			r.prefixBindings[mode][prefix] = true
		}
	} else {
		// Regular keybind - store the action string directly
		r.storeBinding(mode, key, action)

		// Register all prefixes for multi-key commands
		for i := 1; i < len(key); i++ {
			prefix := key[:i]
			r.prefixBindings[mode][prefix] = true
		}
	}

	// Check if binding already exists
	var existingBinding *CommandBinding
	for _, binding := range r.bindings {
		if binding.Key == key {
			existingBinding = binding
			break
		}
	}

	if existingBinding != nil {
		// Add mode to existing binding
		existingBinding.Modes = append(existingBinding.Modes, mode)
	} else {
		// Create new binding for keyhints system
		newBinding := &CommandBinding{
			Key:         key,
			Description: description,
			Modes:       []types.VimMode{mode},
		}
		r.bindings = append(r.bindings, newBinding)
	}
}

// storeBinding is a helper to store action strings in the appropriate binding map
func (r *CommandRegistry) storeBinding(mode types.VimMode, key string, action string) {
	switch mode {
	case types.ModeNormal:
		r.normalBindings[key] = action
	case types.ModeVisual:
		r.visualBindings[key] = action
	case types.ModeSearch:
		r.searchBindings[key] = action
	}
}

// AddFolder registers a command folder/prefix that groups related commands
func (r *CommandRegistry) addFolder(prefix string, description string) {
	// Register the prefix for both modes (folders are usually mode-agnostic)
	r.prefixBindings[types.ModeNormal][prefix] = true
	r.prefixBindings[types.ModeVisual][prefix] = true

	// Store folder description for keyhints system
	newFolder := &CommandFolder{
		Prefix:      prefix,
		Description: description,
	}
	r.folders = append(r.folders, newFolder)
}

// GetBindingsForMode returns all bindings available for the specified mode
func (r *CommandRegistry) GetBindingsForMode(mode types.VimMode) []*CommandBinding {
	var bindings []*CommandBinding
	for _, binding := range r.bindings {
		if slices.Contains(binding.Modes, mode) {
			bindings = append(bindings, binding)
		}
	}
	return bindings
}
