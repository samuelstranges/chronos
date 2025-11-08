package popup

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/samuelstranges/chronos/types"
)

// CalendarSelectorPopup shows a list of calendars with toggle functionality
type CalendarSelectorPopup struct {
	calendars     []types.CalendarInfo
	visibleState  map[string]bool
	selectedIndex int
	onAction      func(CalendarSelectorAction, string) tea.Cmd
}

// CalendarSelectorAction represents available actions in the calendar selector
type CalendarSelectorAction string

const (
	CalendarActionToggle CalendarSelectorAction = "toggle"
	CalendarActionClose  CalendarSelectorAction = "close"
)

var (
	calendarBoxStyle = lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("62")).
				Padding(1, 2).
				Width(50)

	calendarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

	selectedCalendarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)

	normalCalendarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 1)

	visibleCalendarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00bd80")) // Green for visible

	hiddenCalendarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")) // Gray for hidden

	calendarHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Italic(true)
)

// NewCalendarSelectorPopup creates a new calendar selector popup
func NewCalendarSelectorPopup(calendars []types.CalendarInfo, visibleState map[string]bool, onAction func(CalendarSelectorAction, string) tea.Cmd) *CalendarSelectorPopup {
	return &CalendarSelectorPopup{
		calendars:     calendars,
		visibleState:  visibleState,
		selectedIndex: 0,
		onAction:      onAction,
	}
}

// Render implements the Popup interface
func (p *CalendarSelectorPopup) Render(ctx PopupRenderContext) string {
	var content strings.Builder

	content.WriteString(calendarTitleStyle.Render("Calendar Filter"))
	content.WriteString("\n\n")

	if len(p.calendars) == 0 {
		content.WriteString("No calendars found")
	} else {
		content.WriteString(fmt.Sprintf("Toggle calendar visibility (%d calendars):\n\n", len(p.calendars)))

		for i, calendar := range p.calendars {
			calendarLine := p.formatCalendarLine(calendar)

			if i == p.selectedIndex {
				content.WriteString(selectedCalendarStyle.Render("► "))
				content.WriteString(p.getColorPreview(calendar.Color))
				content.WriteString(selectedCalendarStyle.Render(" " + calendarLine))
			} else {
				content.WriteString(normalCalendarStyle.Render("  "))
				content.WriteString(p.getColorPreview(calendar.Color))
				content.WriteString(normalCalendarStyle.Render(" " + calendarLine))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(calendarHelpStyle.Render("↑/↓: navigate • Space/Enter: toggle • Esc: close"))

	return calendarBoxStyle.Render(content.String())
}

// HandleKey implements the Popup interface
func (p *CalendarSelectorPopup) HandleKey(key string) (Popup, tea.Cmd) {
	switch key {
	case "up", "k":
		if len(p.calendars) > 0 && p.selectedIndex > 0 {
			p.selectedIndex--
		}
		return p, nil

	case "down", "j":
		if len(p.calendars) > 0 && p.selectedIndex < len(p.calendars)-1 {
			p.selectedIndex++
		}
		return p, nil

	case "enter", "space": // Space or Enter to toggle
		if len(p.calendars) > 0 && p.onAction != nil {
			selectedCalendar := p.calendars[p.selectedIndex]
			// Update our internal state immediately for responsive UI
			currentState := p.visibleState[selectedCalendar.ID]
			p.visibleState[selectedCalendar.ID] = !currentState
			return p, p.onAction(CalendarActionToggle, selectedCalendar.ID)
		}
		return p, nil

	case "esc", "q":
		if p.onAction != nil {
			return nil, p.onAction(CalendarActionClose, "")
		}
		return nil, nil

	default:
		return p, nil
	}
}

// GetTitle implements the Popup interface
func (p *CalendarSelectorPopup) GetTitle() string {
	return "Calendar Selector"
}

// formatCalendarLine formats a single calendar for display
func (p *CalendarSelectorPopup) formatCalendarLine(calendar types.CalendarInfo) string {
	isVisible := p.visibleState[calendar.ID]

	var statusIcon string
	var nameStyle lipgloss.Style
	if isVisible {
		statusIcon = "✓"
		nameStyle = visibleCalendarStyle
	} else {
		statusIcon = "✗"
		nameStyle = hiddenCalendarStyle
	}

	return fmt.Sprintf("%s %s", statusIcon, nameStyle.Render(calendar.Name))
}

// getColorPreview creates a color preview for the calendar
func (p *CalendarSelectorPopup) getColorPreview(color string) string {
	return CreateColorPreview(color) // Use shared helper
}

// ShowCalendarSelector shows the calendar selector popup
func ShowCalendarSelector(calendars []types.CalendarInfo, visibleState map[string]bool, onAction func(CalendarSelectorAction, string) tea.Cmd) {
	popup := NewCalendarSelectorPopup(calendars, visibleState, onAction)
	ShowPopup(popup)
}

