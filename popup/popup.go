package popup

import tea "github.com/charmbracelet/bubbletea/v2"

// PopupRenderContext contains styling context passed to popups
type PopupRenderContext struct {
	Width           int
	Height          int
	BackgroundColor string
	ForegroundColor string
	BorderColor     string
}

// Popup represents any overlay that can be shown on top of the main UI
type Popup interface {
	// Render returns the popup's visual representation with context
	Render(ctx PopupRenderContext) string
	
	// HandleKey processes keyboard input and returns the next popup state
	// Returns nil to close popup, returns self to stay open, returns different popup to transition
	HandleKey(key string) (Popup, tea.Cmd)
	
	// GetTitle returns the popup title for debugging/logging purposes
	GetTitle() string
}

// PopupManager manages the global popup state
type PopupManager struct {
	current Popup
}

// NewPopupManager creates a new popup manager
func NewPopupManager() *PopupManager {
	return &PopupManager{current: nil}
}

// Show displays a popup, replacing any currently shown popup
func (pm *PopupManager) Show(popup Popup) {
	pm.current = popup
}

// Close removes the current popup
func (pm *PopupManager) Close() {
	pm.current = nil
}

// GetCurrent returns the currently active popup (nil if none)
func (pm *PopupManager) GetCurrent() Popup {
	return pm.current
}


// HandleKey routes keyboard input to the current popup
// Returns (newPopup, cmd) - newPopup can be nil (close), self (stay), or different popup (transition)
func (pm *PopupManager) HandleKey(key string) tea.Cmd {
	if pm.current == nil {
		return nil
	}
	
	newPopup, cmd := pm.current.HandleKey(key)
	pm.current = newPopup // Could be nil (popup closed)
	
	return cmd
}

// Render renders the current popup overlay with context
func (pm *PopupManager) Render(ctx PopupRenderContext) string {
	if pm.current == nil {
		return ""
	}
	
	return pm.current.Render(ctx)
}

// HasPopup returns true if there's an active popup (for UI flow control)
func (pm *PopupManager) HasPopup() bool {
	return pm.current != nil
}

// Global popup manager instance
var GlobalPopupManager = NewPopupManager()

// Convenience functions for global popup management
func ShowPopup(popup Popup) {
	GlobalPopupManager.Show(popup)
}

func ClosePopup() {
	GlobalPopupManager.Close()
}


func HandlePopupKey(key string) tea.Cmd {
	return GlobalPopupManager.HandleKey(key)
}

func RenderPopup(ctx PopupRenderContext) string {
	return GlobalPopupManager.Render(ctx)
}

func HasPopup() bool {
	return GlobalPopupManager.HasPopup()
}