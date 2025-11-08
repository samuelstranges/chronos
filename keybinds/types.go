package keybinds

import (
	"strings"
	"time"

	"github.com/samuelstranges/chronos/types"
)

// VimCommand represents a parsed vim command
type VimCommand struct {
	Count1   int    // First count (5 in "5dw" or "d5w")
	Operator string // d, y, c, v, etc.
	Count2   int    // Second count (5 in "d5w")
	Motion   string // w, j, k, h, l, b, e, etc.
	Text     string // For text objects like "iw", "ap"
}

// VimState manages vim-style key sequence parsing
type VimState struct {
	Mode         types.VimMode              // Current vim mode
	KeySequence  []string                   // Current key sequence being built
	LastKeyTime  time.Time                  // Time of last keypress for timeout
	CountPrefix  int                        // Numeric prefix for commands
	StatusText   string                     // Text to show in status (partial commands)
	VisualAnchor types.CursorPositionInWeek // Starting position of visual selection
}

// CommandBinding represents a complete command binding with metadata (for keyhints)
type CommandBinding struct {
	Key         string
	Description string
	Modes       []types.VimMode
}

// CommandFolder represents a command folder/prefix
type CommandFolder struct {
	Prefix      string
	Description string
}

// CommandRegistry stores key sequences to action strings per mode
type CommandRegistry struct {
	normalBindings map[string]string
	visualBindings map[string]string
	searchBindings map[string]string
	prefixBindings map[types.VimMode]map[string]bool
	patternEntries []PatternEntry // Ordered list of pattern mappings

	// Storage for descriptions and metadata (ordered for deterministic display)
	bindings []*CommandBinding // Ordered list of bindings
	folders  []*CommandFolder  // Ordered list of folders
}

// PatternEntry represents a pattern mapping to an action
type PatternEntry struct {
	Pattern string
	Action  string
}

// NewVimState creates a new vim state manager
func NewVimState() *VimState {
	return &VimState{
		Mode:        types.ModeNormal,
		KeySequence: []string{},
		CountPrefix: 1,
	}
}

// GetStatusText returns the current status text
func (v *VimState) GetStatusText() string {
	return v.StatusText
}

// GetMode returns the current vim mode
func (v *VimState) GetMode() types.VimMode {
	return v.Mode
}

// SetMode changes the current vim mode
func (v *VimState) SetMode(mode types.VimMode) {
	v.Mode = mode
	v.reset() // Clear any pending key sequences
}

// SetVisualAnchor sets the visual selection anchor point
func (v *VimState) SetVisualAnchor(cursor types.CursorPositionInWeek) {
	v.VisualAnchor = cursor
}

// GetVisualAnchor returns the visual selection anchor point
func (v *VimState) GetVisualAnchor() types.CursorPositionInWeek {
	return v.VisualAnchor
}

// reset clears the vim state
func (v *VimState) reset() {
	v.KeySequence = []string{}
	v.CountPrefix = 1
	v.StatusText = ""
}

// updateStatusText updates the status text with current sequence
func (v *VimState) updateStatusText() {
	if len(v.KeySequence) > 0 {
		v.StatusText = strings.Join(v.KeySequence, "")
	} else {
		v.StatusText = ""
	}
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		normalBindings: make(map[string]string),
		visualBindings: make(map[string]string),
		searchBindings: make(map[string]string),
		prefixBindings: map[types.VimMode]map[string]bool{
			types.ModeNormal: make(map[string]bool),
			types.ModeVisual: make(map[string]bool),
			types.ModeSearch: make(map[string]bool),
		},
		patternEntries: make([]PatternEntry, 0),
		bindings:       make([]*CommandBinding, 0),
		folders:        make([]*CommandFolder, 0),
	}

	// Register all keybinds in a centralized way
	registry.registerAllKeybinds()

	return registry
}
