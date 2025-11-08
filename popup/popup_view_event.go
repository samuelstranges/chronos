package popup

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// EventDetailsData holds formatted event information for display
type EventDetailsData struct {
	Title       string
	StartTime   string
	EndTime     string
	Duration    string
	Location    string
	Description string
	Calendar    string
	Color       string
	AllDay      bool
}

// EventDetailsPopup shows read-only event details
type EventDetailsPopup struct {
	event        *ical.Event
	data         *EventDetailsData
	calendarName string
}

// NewEventDetailsPopup creates a new event details popup
func NewEventDetailsPopup(event *ical.Event, calendarName string) *EventDetailsPopup {
	popup := &EventDetailsPopup{
		event:        event,
		calendarName: calendarName,
	}

	// Extract event data
	if event != nil {
		popup.data = popup.extractEventData()
	}

	return popup
}

// Render returns the popup's visual representation
func (p *EventDetailsPopup) Render(ctx PopupRenderContext) string {
	if p.event == nil || p.data == nil {
		return p.renderNoEvent(ctx)
	}

	content := p.formatEventDetails()

	// Style the popup using context dimensions and colors
	popupWidth := ctx.Width * 70 / 100 // 70% of terminal width
	popupStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Background(lipgloss.Color(ctx.BackgroundColor)).
		Foreground(lipgloss.Color(ctx.ForegroundColor)).
		Width(popupWidth).
		Align(lipgloss.Left) // Left align for better readability

	return popupStyle.Render(content)
}

// HandleKey processes keyboard input
func (p *EventDetailsPopup) HandleKey(key string) (Popup, tea.Cmd) {
	switch key {
	case "esc", "q", "enter":
		// Close popup on ESC, q, or Enter
		return nil, nil

	case "o":
		// Open URL if event has one
		if p.event != nil {
			cmd := p.openEventURL()
			return p, cmd // Stay open after opening URL
		}
		return p, nil

	default:
		// For any other key, just stay open
		return p, nil
	}
}

// GetTitle returns the popup title for debugging
func (p *EventDetailsPopup) GetTitle() string {
	if p.data != nil {
		return "EventDetailsPopup: " + p.data.Title
	}
	return "EventDetailsPopup: No Event"
}

// renderNoEvent renders content when no event is selected
func (p *EventDetailsPopup) renderNoEvent(ctx PopupRenderContext) string {
	content := "No event selected"

	popupStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Background(lipgloss.Color(ctx.BackgroundColor)).
		Foreground(lipgloss.Color(ctx.ForegroundColor)).
		Width(30).
		Align(lipgloss.Center)

	return popupStyle.Render(content)
}

// extractEventData extracts formatted data from the iCal event using utility functions
func (p *EventDetailsPopup) extractEventData() *EventDetailsData {
	data := &EventDetailsData{}

	// Title with fallback
	data.Title = util.GetEventTitle(p.event)
	if data.Title == "" {
		data.Title = "Untitled Event"
	}

	data.Location = util.GetEventLocation(p.event)
	data.Description = util.GetEventDescription(p.event)
	data.Color = util.GetEventColor(p.event, "")
	data.Calendar = p.calendarName
	data.AllDay = util.IsAllDayEvent(p.event)
	startTime, endTime, err := timezone.GetEventTimes(p.event)
	if err != nil {
		// Handle error gracefully - set zero times
		startTime = timezone.LocalTime{}
		endTime = timezone.LocalTime{}
	}

	if !startTime.IsZero() {
		if data.AllDay {
			p.formatAllDayTimes(data, startTime.Time, endTime.Time)
		} else {
			p.formatTimedEventTimes(data, startTime.Time, endTime.Time)
		}
	}

	return data
}

// formatAllDayTimes formats time information for all-day events
func (p *EventDetailsPopup) formatAllDayTimes(data *EventDetailsData, startTime, endTime time.Time) {
	startTimeLocal := startTime.In(time.Local)
	data.StartTime = startTimeLocal.Format("Mon Jan 2, 2006")

	if !endTime.IsZero() {
		endTimeLocal := endTime.In(time.Local)
		// For all-day events, DTEND is exclusive, so subtract a second for display
		data.EndTime = endTimeLocal.Add(-1 * time.Second).Format("Mon Jan 2, 2006")

		data.Duration = endTime.Sub(startTime).String()
	} else {
		data.EndTime = data.StartTime
		data.Duration = "1 day"
	}
}

// formatTimedEventTimes formats time information for timed events
func (p *EventDetailsPopup) formatTimedEventTimes(data *EventDetailsData, startTime, endTime time.Time) {
	startTimeLocal := startTime.In(time.Local)
	data.StartTime = startTimeLocal.Format("Mon Jan 2, 2006 at 3:04 PM")

	if !endTime.IsZero() {
		endTimeLocal := endTime.In(time.Local)
		data.EndTime = endTimeLocal.Format("Mon Jan 2, 2006 at 3:04 PM")
		data.Duration = endTime.Sub(startTime).String()
	}
}

// formatIfExists returns formatted string with newline prefix if value is not empty
func formatIfExists(value, format string, args ...interface{}) string {
	if value != "" {
		if len(args) > 0 {
			return "\n\n" + fmt.Sprintf(format, args...)
		} else {
			return "\n\n" + fmt.Sprintf(format, value)
		}
	}
	return ""
}

// formatTimeInformation returns the time section (all-day vs timed events)
func (p *EventDetailsPopup) formatTimeInformation() string {
	if p.data.AllDay {
		result := fmt.Sprintf("\n\nType: All Day - %s", p.data.Duration)
		if p.data.StartTime == p.data.EndTime {
			result += fmt.Sprintf("\nDate: %s", p.data.StartTime)
		} else {
			result += fmt.Sprintf("\nDates: %s to %s", p.data.StartTime, p.data.EndTime)
		}
		return result
	} else {
		return fmt.Sprintf("\n\nDuration: %s\nStarts: %s\nEnds: %s",
			p.data.Duration, p.data.StartTime, p.data.EndTime)
	}
}

// formatEventDetails formats the event data for display
func (p *EventDetailsPopup) formatEventDetails() string {
	colorPreview := CreateColorPreview(p.data.Color)

	result := "📅 EVENT DETAILS\n\n"
	result += fmt.Sprintf("Title: %s", p.data.Title)
	result += p.formatTimeInformation()
	result += formatIfExists(p.data.Location, "Location: %s")
	result += formatIfExists(p.data.Calendar, "Calendar: %s")
	result += formatIfExists(p.data.Color, "Color: %s %s", p.data.Color, colorPreview)
	result += formatIfExists(p.data.Description, "Description:\n%s")
	result += "\n\nPress ESC/Enter to close, 'o' to open URL"

	return result
}


// openEventURL opens the event's URL if it has one
func (p *EventDetailsPopup) openEventURL() tea.Cmd {
	if p.event == nil {
		return func() tea.Msg {
			return types.VimErrorMsg{Error: "No event to open URL for"}
		}
	}

	// Extract URL using utility function
	eventURL := util.GetEventLink(p.event)
	if eventURL == "" {
		return func() tea.Msg {
			return types.VimErrorMsg{Error: "Event has no URL"}
		}
	}

	// Return command to open URL (delegate to existing system)
	return func() tea.Msg {
		return types.VimActionMsg{Action: "open_url"}
	}
}

// ShowEventDetailsPopup is a convenience function to show event details popup
func ShowEventDetailsPopup(event *ical.Event, calendarName string) {
	popup := NewEventDetailsPopup(event, calendarName)
	ShowPopup(popup)
}
