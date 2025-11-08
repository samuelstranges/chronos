package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samuelstranges/chronos/keybinds_hints"
	types "github.com/samuelstranges/chronos/types"
)

// ActionHandler defines the interface for all vim actions
type ActionHandler func(m Model, argument int) (tea.Model, tea.Cmd)

var actionHandlers = map[string]ActionHandler{
	// Application Control
	"quit": HandleQuit,

	// Hints
	"toggle_hints":              HandleToggleHints,
	"hide_hints":                HandleHideHints,
	"update_hints_for_sequence": HandleUpdateHintsForSequence,
	"start_hint_timeout":        HandleStartHintTimeout,

	// Forms
	"close_form": HandleCloseForm,

	// Clipboard Operations
	"cut":   HandleCutEvent,
	"copy":  HandleCopyEvent,
	"paste": HandlePasteEvent,

	// Undo/Redo
	"undo": HandleUndoAction,
	"redo": HandleRedoAction,

	// View Layers
	"toggle_agenda":       HandleToggleAgendaLayer,
	"toggle_all_day_grid": HandleToggleAllDayGrid,
	"toggle_event_text_color": HandleToggleEventTextColor,

	// Event Actions
	"add_event":              HandleAddEvent,
	"edit_event":             HandleEditEvent,
	"view_event_details":     HandleViewEventDetails,
	"open_url":               HandleOpenEventURL,
	"open_all_day_event_url": HandleOpenAllDayEventURL,

	// Calendar
	"calendar_popup": HandleCalendarPopup,

	// Change popups
	"change_color":       HandleChangeColor,
	"change_duration":    HandleChangeDuration,
	"change_description": HandleChangeDescription,
	"change_start":       HandleChangeStartTime,
	"change_end":         HandleChangeEndTime,
	"change_title":       HandleChangeTitle,
	"change_location":    HandleChangeLocation,
	"change_link":        HandleChangeLink,

	// Visual Mode
	"enter_visual":        HandleEnterVisualMode,
	"exit_visual":         HandleExitVisualMode,
	"copy_selection":      HandleCopySelection,
	"cut_selection":       HandleCutSelection,
	"delete_selection":    HandleDeleteSelection,
	"swap_selection_ends": HandleSwapSelectionEnds,

	// Search Mode
	"enter_search_mode":   HandleEnterSearchMode,
	"exit_search_mode":    HandleExitSearchMode,
	"execute_search":      HandleExecuteSearch,
	"search_next":         HandleNavigateToNextResult,
	"search_previous":     HandleNavigateToPreviousResult,
	"clear_locked_search": HandleClearLockedSearch,

	// Navigation
	"move_to_next_event":     HandleNavigateToNextEvent,
	"move_to_previous_event": HandleNavigateToPreviousEvent,
	"move_to_end_of_event":   HandleNavigateToEventEnd, // 'e' key binding

	// Movement (shared between normal and visual modes)
	"move_left":                  HandleMoveLeft,
	"move_right":                 HandleMoveRight,
	"move_up":                    HandleMoveUp,
	"move_down":                  HandleMoveDown,
	"visual_move_left":           HandleMoveLeft,
	"visual_move_right":          HandleMoveRight,
	"visual_move_up":             HandleMoveUp,
	"visual_move_down":           HandleMoveDown,
	"navigate_to_next_week":      HandleNavigateToNextWeek,
	"navigate_to_previous_week":  HandleNavigateToPreviousWeek,
	"jump_to_current_time":       HandleJumpToCurrentTime,
	"jump_to_end_of_day":         HandleJumpToEndOfDay,
	"jump_to_start_of_day":       HandleJumpToStartOfDay,
	"jump_to_first_event_of_day": HandleJumpToFirstEventOfDay,
	"jump_to_last_event_of_day":  HandleJumpToLastEventOfDay,

	// Zoom
	"zoom_in":  HandleZoomIn,
	"zoom_out": HandleZoomOut,

	// Recurring Events
	"show_recurring_delete_dialog":             HandleShowRecuringDeleteDialog,
	"show_recurring_delete_dialog_from_picker": HandleRecurringDeleteFromPicker,

	// Navigation pattern commands
	"go_to_hour":   HandleGoToHour,
	"go_to_minute": HandleGoToMinute,
	"go_to_time":   HandleGoToTime,
	"go_to_day":    HandleGoToDay,
	"go_to_month":  HandleGoToMonth,
	"go_to_year":   HandleGoToYear,
}

// handleVimAction processes vim action messages using the action handler registry
func handleVimAction(m Model, msg types.VimActionMsg) (tea.Model, tea.Cmd) {
	// Default argument to 1 if not specified (for repeat counts)
	argument := msg.Argument
	if argument <= 0 {
		argument = 1
	}

	// Check if we have a registered handler for this action
	if handler, exists := actionHandlers[msg.Action]; exists {
		return handler(m, argument)
	}

	// Handle navigate_hints actions as fallback
	if keybinds_hints.HandleNavigateHints(m.KeyHints, msg.Action) {
		return m, nil
	}

	return m, nil
}
