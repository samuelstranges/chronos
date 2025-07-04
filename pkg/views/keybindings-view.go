package views

import (
	"fmt"
	"github.com/jroimartin/gocui"
)

type KeybindsView struct {
	*BaseView
	IsVisible bool
}

func NewKeybindsView() *KeybindsView {
	return &KeybindsView{
		BaseView:  NewBaseView("keybinds"),
		IsVisible: false,
	}
}

// getKeybindingsContent returns the keybindings content as a slice of strings
func (kbv *KeybindsView) getKeybindingsContent() []string {
	return []string{
		"",
		" q           - Quit",
		"",
		" Navigation:",
		" h/l or ←/→  - Previous/Next day",
		" H/L         - Previous/Next week",
		" m/M         - Previous/Next month",
		" j/k or ↓/↑  - Move time cursor down/up",
		" t           - Jump to today",
		" T           - To date/time form",
		" w/b         - Jump to next/prev event",
		" e           - End of current/next event",
		" g/G         - Start/End of day (00:00/23:30)",
		"",
		" Views:",
		" v           - Toggle view (Week → Month → Agenda)",
		"",
		"",
		" Advanced Search:",
		" /           - Search events (name/desc/loc)",
		" n           - Next search match",
		" N           - Previous search match",
		" Esc         - Clear search",
		"",
		" Event Management:",
		" a           - Add new event",
		" c           - Change event",
		" C           - Color picker",
		" y           - Copy event",
		" p           - Paste event",
		" d           - Delete event",
		" D           - Delete all events w/ same name",
		"",
		" Undo Buffer:",
		" u           - Undo last action",
		" r           - Redo last undone action",
	}
}

// GetRequiredHeight returns the number of lines needed for all keybinding content
func (kbv *KeybindsView) GetRequiredHeight() int {
	lines := kbv.getKeybindingsContent()
	return len(lines) + 2 // +2 for top and bottom borders
}
func (kbv *KeybindsView) Update(g *gocui.Gui) error {
	if !kbv.IsVisible {
		return nil
	}
	v, err := g.SetView(
		kbv.Name,
		kbv.X,
		kbv.Y,
		kbv.X+kbv.W,
		kbv.Y+kbv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Keybindings "
	}
	v.Clear()
	lines := kbv.getKeybindingsContent()
	for _, line := range lines {
		fmt.Fprintln(v, line)
	}
	g.SetViewOnTop("keybinds")
	return nil
}
