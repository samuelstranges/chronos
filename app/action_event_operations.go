package app

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/popup"
	"github.com/samuelstranges/chronos/types"
	util "github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// openICalEventURL opens the URL of a specific ical.Event
func openICalEventURL(event *ical.Event) error {
	if event == nil {
		return fmt.Errorf("no event provided")
	}

	// Extract URL property
	urlProp := event.Props.Get("URL")
	if urlProp == nil || urlProp.Value == "" {
		return fmt.Errorf("event has no URL")
	}

	// Execute open command
	cmd := exec.Command("open", urlProp.Value)
	return cmd.Start()
}

func HandleViewEventDetails(m Model, count int) (tea.Model, tea.Cmd) {
	// Get event at cursor position
	selectedEvent := week_view_grid.GetEventUnderCursor(&m.WeekModel)

	var icalEvent *ical.Event = nil
	calendarName := "Unknown Calendar"

	if selectedEvent != nil && selectedEvent.Instance != nil && selectedEvent.Instance.OriginalEvent != nil {
		icalEvent = selectedEvent.Instance.OriginalEvent
		// Get calendar name from EventManager
		calendarName = m.EventManager.GetCalendarNameForEvent(icalEvent)
	}

	// Show new popup system event details
	popup.ShowEventDetailsPopup(icalEvent, calendarName)

	return m, nil
}

// HandleOpenEventURL opens the URL of the event at cursor position
func HandleOpenEventURL(m Model, count int) (tea.Model, tea.Cmd) {
	// Get the event at cursor position
	selectedEvent := week_view_grid.GetEventUnderCursor(&m.WeekModel)
	if selectedEvent == nil || selectedEvent.Instance == nil || selectedEvent.Instance.OriginalEvent == nil {
		util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, fmt.Errorf("no event at cursor position"), "Cant show url")
		return m, nil
	}

	err := openICalEventURL(selectedEvent.Instance.OriginalEvent)
	util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Cant show url")
	return m, nil
}

// HandleOpenAllDayEventURL opens the URL of the selected all-day event
func HandleOpenAllDayEventURL(m Model, count int) (tea.Model, tea.Cmd) {
	if m.WeekModel.SelectedAllDayEvent != nil {
		err := openICalEventURL(m.WeekModel.SelectedAllDayEvent)
		util.ShowIfError(&m.WeekModel, types.ErrorDisplaySeconds, err, "Cant show url")

		// Clear the stored event after use
		m.WeekModel.SelectedAllDayEvent = nil
	}
	return m, nil
}
