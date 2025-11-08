package popup_better_forms

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/huh/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/popup"
)

// FormPopup is a generic popup that can display any huh.Form
type FormPopup struct {
	form     *huh.Form
	title    string
	onSubmit func() tea.Cmd // Called when form is completed
	onCancel func() tea.Cmd // Called when user presses ESC
}

// NewFormPopup creates a new generic form popup
func NewFormPopup(form *huh.Form, title string, onSubmit, onCancel func() tea.Cmd) *FormPopup {
	// Initialize the form so it can render properly
	form.Init()

	return &FormPopup{
		form:     form,
		title:    title,
		onSubmit: onSubmit,
		onCancel: onCancel,
	}
}

// Render implements popup.Popup interface
func (p *FormPopup) Render(ctx popup.PopupRenderContext) string {
	if p.form == nil {
		return "No form available"
	}

	// Set form width to available context width (with some margin for styling)
	formWidth := ctx.Width - 10 // Leave margin for padding and borders
	if formWidth < 50 {
		formWidth = 50 // Minimum width
	}
	p.form.WithWidth(formWidth)
	
	// Reinitialize the form with the new width so placeholders render correctly
	p.form.Init()

	// Render the form
	formContent := p.form.View()

	// Calculate popup width (80% of terminal width, max 100 chars)
	popupWidth := ctx.Width * 80 / 100
	popupWidth = min(popupWidth, 100)
	popupWidth = max(popupWidth, 50)

	// Style the popup with context colors (no border - popup system handles that)
	popupStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Background(lipgloss.Color(ctx.BackgroundColor)).
		Foreground(lipgloss.Color(ctx.ForegroundColor)).
		Width(popupWidth)

	return popupStyle.Render(formContent)
}

// HandleKey implements popup.Popup interface
// FormPopup key handling is now done directly by main app - this is just for compatibility
func (p *FormPopup) HandleKey(key string) (popup.Popup, tea.Cmd) {
	// Note: This method is no longer used - main app handles form input directly
	// Only kept for popup.Popup interface compatibility
	// All form handling happens in app/update.go handleFormPopupUpdate()
	return p, nil
}

// GetTitle implements popup.Popup interface
func (p *FormPopup) GetTitle() string {
	if p.title != "" {
		return p.title
	}
	return "Form"
}

// GetForm returns the huh form for main app access
func (p *FormPopup) GetForm() *huh.Form {
	return p.form
}

// SetForm updates the huh form (called by main app after Update)
func (p *FormPopup) SetForm(form *huh.Form) {
	p.form = form
}

// IsCompleted checks if the form is completed
func (p *FormPopup) IsCompleted() bool {
	return p.form != nil && p.form.State == huh.StateCompleted
}

// OnSubmit returns the submit callback for main app to call
func (p *FormPopup) OnSubmit() func() tea.Cmd {
	return p.onSubmit
}

// Convenience functions for showing form popups

// ShowFormPopup shows a form popup using the global popup manager
func ShowFormPopup(form *huh.Form, title string, onSubmit, onCancel func() tea.Cmd) {
	formPopup := NewFormPopup(form, title, onSubmit, onCancel)
	popup.ShowPopup(formPopup)
}


