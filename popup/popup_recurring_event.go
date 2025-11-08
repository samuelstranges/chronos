package popup

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	util "github.com/samuelstranges/chronos/util"
)

// RecurringDialogChoice represents the user's choice for recurring operations
type RecurringDialogChoice string

const (
	ChoiceThis   RecurringDialogChoice = "this"
	ChoiceFuture RecurringDialogChoice = "future"
	ChoiceAll    RecurringDialogChoice = "all"
	ChoiceCancel RecurringDialogChoice = "cancel"
)

// RecurringDialogPopup shows the "This/Future/All" choice dialog for recurring events
type RecurringDialogPopup struct {
	event         *ical.Event
	operationType types.RecurringEventOperationType
	selectedIndex int
	eventTitle    string
	actionText    string
	onChoice      func(RecurringDialogChoice) tea.Cmd
}

var (
	dialogBoxStyle      = lipgloss.NewStyle().BorderForeground(lipgloss.Color("62")).Padding(1, 2).Width(50)
	titleStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedOptionStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	normalOptionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	warningStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

// NewRecurringDialogPopup creates a new recurring dialog popup
func NewRecurringDialogPopup(event *ical.Event, operationType types.RecurringEventOperationType, onChoice func(RecurringDialogChoice) tea.Cmd) *RecurringDialogPopup {
	eventTitle := util.GetEventTitle(event)
	if eventTitle == "" {
		eventTitle = "Untitled Event"
	}

	var actionText string
	switch operationType {
	case types.OperationTypeDelete:
		actionText = "delete"
	case types.OperationTypeEdit:
		actionText = "edit"
	default:
		actionText = "modify"
	}

	return &RecurringDialogPopup{
		event:         event,
		operationType: operationType,
		selectedIndex: 0,
		eventTitle:    eventTitle,
		actionText:    actionText,
		onChoice:      onChoice,
	}
}

// Render implements the Popup interface
func (p *RecurringDialogPopup) Render(ctx PopupRenderContext) string {
	title := p.getDialogTitle()

	// Create the content
	var content strings.Builder
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Event: %s\n", p.eventTitle))
	content.WriteString(fmt.Sprintf("This event repeats. Choose what to %s:\n\n", p.actionText))

	// Add options
	options := []string{
		"This event only",
		"This and all future events",
		"All events in the series",
		"Cancel",
	}

	for i, option := range options {
		if i == p.selectedIndex {
			content.WriteString(selectedOptionStyle.Render("► " + option))
		} else {
			content.WriteString(normalOptionStyle.Render("  " + option))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(warningStyle.Render("Warning: This action cannot be undone"))
	content.WriteString("\n\n")
	content.WriteString("Use ↑/↓ to navigate, Enter to select, Esc to cancel")

	return dialogBoxStyle.Render(content.String())
}

// HandleKey implements the Popup interface
func (p *RecurringDialogPopup) HandleKey(key string) (Popup, tea.Cmd) {
	switch key {
	case "up", "k":
		if p.selectedIndex > 0 {
			p.selectedIndex--
		}
		return p, nil

	case "down", "j":
		if p.selectedIndex < 3 { // 4 options (0-3)
			p.selectedIndex++
		}
		return p, nil

	case "enter":
		choice := p.getChoiceFromIndex()
		if p.onChoice != nil {
			return nil, p.onChoice(choice) // Close popup and execute choice
		}
		return nil, nil

	case "esc", "q":
		if p.onChoice != nil {
			return nil, p.onChoice(ChoiceCancel)
		}
		return nil, nil

	default:
		return p, nil
	}
}

// GetTitle implements the Popup interface
func (p *RecurringDialogPopup) GetTitle() string {
	return p.getDialogTitle()
}

// getDialogTitle returns the dialog title based on operation type
func (p *RecurringDialogPopup) getDialogTitle() string {
	switch p.operationType {
	case types.OperationTypeDelete:
		return "Delete Recurring Event"
	case types.OperationTypeEdit:
		return "Edit Recurring Event"
	default:
		return "Recurring Event Operation"
	}
}

// getChoiceFromIndex converts selected index to choice
func (p *RecurringDialogPopup) getChoiceFromIndex() RecurringDialogChoice {
	switch p.selectedIndex {
	case 0:
		return ChoiceThis
	case 1:
		return ChoiceFuture
	case 2:
		return ChoiceAll
	case 3:
		return ChoiceCancel
	default:
		return ChoiceCancel
	}
}

// ShowRecurringDialog shows the recurring dialog popup
func ShowRecurringDialog(event *ical.Event, operationType types.RecurringEventOperationType, onChoice func(RecurringDialogChoice) tea.Cmd) {
	popup := NewRecurringDialogPopup(event, operationType, onChoice)
	ShowPopup(popup)
}
