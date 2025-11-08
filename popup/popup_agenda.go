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

// AgendaPickerPopup shows a list of events for a specific day
type AgendaPickerPopup struct {
	events        []*types.EventInstance
	selectedIndex int
	targetDay     int
	dayName       string
	onAction      func(AgendaPickerAction, *ical.Event) tea.Cmd
}

// AgendaPickerAction represents available actions in the agenda picker
type AgendaPickerAction string

const (
	AgendaActionViewDetails AgendaPickerAction = "view_details"
	AgendaActionEdit        AgendaPickerAction = "edit"
	AgendaActionDelete      AgendaPickerAction = "delete"
	AgendaActionOpenURL     AgendaPickerAction = "open_url"
	AgendaActionClose       AgendaPickerAction = "close"
)

var (
	agendaBoxStyle     = lipgloss.NewStyle().BorderForeground(lipgloss.Color("62")).Padding(types.PaddingVerticalStandard, types.PaddingHorizontalStandard).Width(types.PopupWidthLarge)
	agendaTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedEventStyle = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	normalEventStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
)

// NewAgendaPickerPopup creates a new agenda picker popup
func NewAgendaPickerPopup(events []*types.EventInstance, targetDay int, dayName string, onAction func(AgendaPickerAction, *ical.Event) tea.Cmd) *AgendaPickerPopup {
	return &AgendaPickerPopup{
		events:        events,
		selectedIndex: 0,
		targetDay:     targetDay,
		dayName:       dayName,
		onAction:      onAction,
	}
}

// Render implements the Popup interface
func (p *AgendaPickerPopup) Render(ctx PopupRenderContext) string {
	var content strings.Builder

	title := fmt.Sprintf("Agenda - %s", p.dayName)
	content.WriteString(agendaTitleStyle.Render(title))
	content.WriteString("\n\n")

	if len(p.events) == 0 {
		content.WriteString("No events for this day")
	} else {
		content.WriteString(fmt.Sprintf("Events (%d):\n\n", len(p.events)))

		for i, event := range p.events {
			colorPreview, timeAndTitle := p.formatEventLineParts(event)

			if i == p.selectedIndex {
				// Style each part separately - arrow and text get blue background, color preview doesn't
				content.WriteString(selectedEventStyle.Render("► "))
				content.WriteString(colorPreview)
				content.WriteString(selectedEventStyle.Render(" " + timeAndTitle))
			} else {
				// For unselected items, style arrow and text normally, color preview is unstyled
				content.WriteString(normalEventStyle.Render("  "))
				content.WriteString(colorPreview)
				content.WriteString(normalEventStyle.Render(" " + timeAndTitle))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑/↓: navigate • Enter: view • c: edit • x: delete • o: open URL • Esc: close"))

	return agendaBoxStyle.Render(content.String())
}

// HandleKey implements the Popup interface
func (p *AgendaPickerPopup) HandleKey(key string) (Popup, tea.Cmd) {
	switch key {
	case "up", "k":
		if len(p.events) > 0 && p.selectedIndex > 0 {
			p.selectedIndex--
		}
		return p, nil

	case types.KeyDown, "j":
		if len(p.events) > 0 && p.selectedIndex < len(p.events)-1 {
			p.selectedIndex++
		}
		return p, nil

	case types.KeyEnter:
		if len(p.events) > 0 && p.onAction != nil {
			selectedEvent := p.events[p.selectedIndex].OriginalEvent
			return nil, p.onAction(AgendaActionViewDetails, selectedEvent)
		}
		return nil, nil

	case "c":
		if len(p.events) > 0 && p.onAction != nil {
			selectedEvent := p.events[p.selectedIndex].OriginalEvent
			return nil, p.onAction(AgendaActionEdit, selectedEvent)
		}
		return p, nil

	case "x":
		if len(p.events) > 0 && p.onAction != nil {
			selectedEvent := p.events[p.selectedIndex].OriginalEvent
			return nil, p.onAction(AgendaActionDelete, selectedEvent)
		}
		return p, nil

	case "o":
		if len(p.events) > 0 && p.onAction != nil {
			selectedEvent := p.events[p.selectedIndex].OriginalEvent
			return p, p.onAction(AgendaActionOpenURL, selectedEvent) // Keep popup open for URL
		}
		return p, nil

	case types.KeyEsc, "q":
		if p.onAction != nil {
			return nil, p.onAction(AgendaActionClose, nil)
		}
		return nil, nil

	default:
		return p, nil
	}
}

// GetTitle implements the Popup interface
func (p *AgendaPickerPopup) GetTitle() string {
	return fmt.Sprintf("Agenda Picker - %s", p.dayName)
}

// formatEventLineParts formats a single event into separate color and text parts
func (p *AgendaPickerPopup) formatEventLineParts(event *types.EventInstance) (string, string) {
	title := util.GetEventTitle(event.OriginalEvent)
	if title == "" {
		title = types.UntitledEvent
	}

	// Get event color and create color preview (unstyled by background)
	eventColor := util.GetEventColor(event.OriginalEvent, "")
	colorPreview := CreateColorPreview(eventColor)

	// Format time
	var timeStr string
	if util.IsAllDayEvent(event.OriginalEvent) {
		timeStr = "All Day    " // with extra spacing to align with time
	} else {
		startTime := event.ComputedStart.Format("15:04")
		endTime := event.ComputedEnd.Format("15:04")
		timeStr = fmt.Sprintf("%s-%s", startTime, endTime)
	}

	// Return color preview separately from time and title
	timeAndTitle := fmt.Sprintf("%s - %s", timeStr, title)
	return colorPreview, timeAndTitle
}

// ShowAgendaPicker shows the agenda picker popup
func ShowAgendaPicker(events []*types.EventInstance, targetDay int, dayName string, onAction func(AgendaPickerAction, *ical.Event) tea.Cmd) {
	popup := NewAgendaPickerPopup(events, targetDay, dayName, onAction)
	ShowPopup(popup)
}
