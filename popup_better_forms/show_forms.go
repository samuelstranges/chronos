package popup_better_forms

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_grid"
)

// Helper function to update event time property with proper formatting
func updateEventTimeProperty(event *ical.Event, originalTime timezone.LocalTime, newTimeStr, propertyName string) (map[string]string, error) {
	// Parse the new time (HH:MM format)
	newTime, err := time.ParseInLocation("15:04", newTimeStr, originalTime.Time.Location())
	if err != nil || originalTime.IsZero() {
		return nil, err
	}
	
	// Create new time with same date but new time
	newDateTime := time.Date(
		originalTime.Time.Year(), originalTime.Time.Month(), originalTime.Time.Day(),
		newTime.Hour(), newTime.Minute(), 0, 0,
		originalTime.Time.Location(),
	)
	
	// Format the time string
	var timeStr string
	if util.IsAllDayEvent(event) {
		timeStr = newDateTime.Format("20060102")
	} else {
		// Use utility function to preserve original timezone format
		originalProp := event.Props.Get(propertyName)
		originalValue := ""
		if originalProp != nil {
			originalValue = originalProp.Value
		}
		timeStr = timezone.PreserveTimezoneFormat(timezone.LocalTime{Time: newDateTime}, originalValue)
	}

	// Validate the time range using existing utility
	var otherTimeStr string
	if propertyName == "DTSTART" {
		// We're updating start time, get current end time for validation
		if endProp := event.Props.Get("DTEND"); endProp != nil {
			otherTimeStr = endProp.Value
		}
		if !util.IsValidTimeRange(timeStr, otherTimeStr) {
			return nil, fmt.Errorf("start time must be before end time")
		}
	} else if propertyName == "DTEND" {
		// We're updating end time, get current start time for validation
		if startProp := event.Props.Get("DTSTART"); startProp != nil {
			otherTimeStr = startProp.Value
		}
		if !util.IsValidTimeRange(otherTimeStr, timeStr) {
			return nil, fmt.Errorf("end time must be after start time")
		}
	}

	return map[string]string{
		propertyName:     timeStr,
		"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
	}, nil
}

// EventUpdateHandler defines the interface for event property update handlers
type EventUpdateHandler func(properties map[string]string) tea.Cmd

// Example usage functions showing how clean the API is

// Simple single-field edit functions

// ShowEditTitle shows a popup to edit an event's title
func ShowEditTitle(event *ical.Event, onUpdate EventUpdateHandler) {
	title := util.GetEventTitle(event)
	form := SimpleTitleForm(&title)

	ShowFormPopup(form, "Edit Title", func() tea.Cmd {
		properties := map[string]string{
			"SUMMARY":       title,
			"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
		}
		return onUpdate(properties)
	}, nil)
}

// ShowEditColor shows a popup to edit an event's color
func ShowEditColor(event *ical.Event, onUpdate EventUpdateHandler) {
	color := util.GetEventColor(event, "")
	form := SimpleColorForm(&color)

	ShowFormPopup(form, "Edit Color", func() tea.Cmd {
		// Transform shorthand color to hex code
		hexColor := util.ParseColor(color)
		if hexColor == "" {
			hexColor = color // Fallback to original if parsing fails
		}

		properties := map[string]string{
			"X-COLOR":       hexColor,
			"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
		}
		return onUpdate(properties)
	}, nil)
}

// ShowEditDuration shows a popup to edit an event's duration
func ShowEditDuration(event *ical.Event, onUpdate EventUpdateHandler) {
	// Calculate current duration in minutes
	startTime, endTime, err := timezone.GetEventTimes(event)
	if err != nil {
		// Handle error gracefully - use zero times and default duration
		startTime = timezone.LocalTime{}
		endTime = timezone.LocalTime{}
	}
	duration := "60" // Default
	if !startTime.IsZero() && !endTime.IsZero() {
		duration = fmt.Sprintf("%.0f", endTime.Time.Sub(startTime.Time).Minutes())
	}

	form := SimpleDurationForm(&duration)

	ShowFormPopup(form, "Edit Duration", func() tea.Cmd {
		// Parse new duration and calculate new end time
		if durationMins, err := strconv.Atoi(duration); err == nil && durationMins > 0 && !startTime.IsZero() {
			newEndTime := startTime.Add(time.Duration(durationMins) * time.Minute)

			var endTimeStr string
			if util.IsAllDayEvent(event) {
				endTimeStr = newEndTime.Time.Format("20060102")
			} else {
				// Use utility function to preserve original timezone format
				originalDtend := event.Props.Get("DTEND")
				originalDtendValue := ""
				if originalDtend != nil {
					originalDtendValue = originalDtend.Value
				}
				endTimeStr = timezone.PreserveTimezoneFormat(newEndTime, originalDtendValue)
			}

			properties := map[string]string{
				"DTEND":         endTimeStr,
				"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
			}
			return onUpdate(properties)
		}
		return nil // Invalid duration, do nothing
	}, nil)
}

// ShowEditDescription shows a popup to edit an event's description
func ShowEditDescription(event *ical.Event, onUpdate EventUpdateHandler) {
	description := util.GetEventDescription(event)
	form := SimpleDescriptionForm(&description)

	ShowFormPopup(form, "Edit Description", func() tea.Cmd {
		properties := map[string]string{
			"DESCRIPTION":   description,
			"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
		}
		return onUpdate(properties)
	}, nil)
}

// ShowEditLocation shows a popup to edit an event's location
func ShowEditLocation(event *ical.Event, onUpdate EventUpdateHandler) {
	location := util.GetEventLocation(event)
	form := SimpleLocationForm(&location)

	ShowFormPopup(form, "Edit Location", func() tea.Cmd {
		properties := map[string]string{
			"LOCATION":      location,
			"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
		}
		return onUpdate(properties)
	}, nil)
}

// ShowEditLink shows a popup to edit an event's link/URL
func ShowEditLink(event *ical.Event, onUpdate EventUpdateHandler) {
	link := util.GetEventLink(event)
	form := SimpleLinkForm(&link)

	ShowFormPopup(form, "Edit Link", func() tea.Cmd {
		properties := map[string]string{
			"URL":           link,
			"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
		}
		return onUpdate(properties)
	}, nil)
}

// ShowEditStartTime shows a popup to edit an event's start time
func ShowEditStartTime(event *ical.Event, onUpdate EventUpdateHandler) {
	// Get current start time and format as HH:MM
	startTime, _, err := timezone.GetEventTimes(event)
	timeStr := "09:00" // Default
	if err == nil && !startTime.IsZero() {
		timeStr = startTime.Time.Format("15:04")
	}
	
	form := SimpleStartTimeForm(&timeStr)

	ShowFormPopup(form, "Edit Start Time", func() tea.Cmd {
		// Use helper function to update start time
		if properties, err := updateEventTimeProperty(event, startTime, timeStr, "DTSTART"); err == nil {
			return onUpdate(properties)
		} else {
			// Return error message for status bar display
			return func() tea.Msg {
				return types.VimErrorMsg{Error: err.Error()}
			}
		}
	}, nil)
}

// ShowEditEndTime shows a popup to edit an event's end time
func ShowEditEndTime(event *ical.Event, onUpdate EventUpdateHandler) {
	// Get current end time and format as HH:MM
	_, endTime, err := timezone.GetEventTimes(event)
	timeStr := "10:00" // Default
	if err == nil && !endTime.IsZero() {
		timeStr = endTime.Time.Format("15:04")
	}
	
	form := SimpleEndTimeForm(&timeStr)

	ShowFormPopup(form, "Edit End Time", func() tea.Cmd {
		// Use helper function to update end time
		if properties, err := updateEventTimeProperty(event, endTime, timeStr, "DTEND"); err == nil {
			return onUpdate(properties)
		} else {
			// Return error message for status bar display
			return func() tea.Msg {
				return types.VimErrorMsg{Error: err.Error()}
			}
		}
	}, nil)
}

// Event creation function following popup_better_forms patterns

// EventCreateHandler defines the interface for event creation handlers
type EventCreateHandler func(properties map[string]string, startTime, endTime time.Time, isAllDay bool, calendarID string) tea.Cmd

// ShowAddEvent shows a popup to create a new event
func ShowAddEvent(weekModel *types.WeekModel, calendars []types.CalendarInfo, onCreate EventCreateHandler) {
	// Initialize form variables with defaults from cursor position
	startDateTime := week_view_grid.CellToTimeAtCursor(*weekModel)
	weekLocation := weekModel.CurrentlyViewedWeek.Location()

	// Form variables - following SimpleAddEventForm parameters
	title := ""
	eventType := "timed" // Default to timed events
	color := ""          // Use calendar default
	calendarID := ""     // Will default to first calendar
	startDate := startDateTime.Format("2006-01-02")
	startTime := startDateTime.Format("15:04")
	duration := "60"     // Default 1 hour
	endDate := startDate // Same day for all-day events
	description := ""
	location := ""
	link := ""
	formComplexity := "simple"   // Start simple
	eventRecurrence := "onetime" // Default to one-time
	recurrenceType := "weekly"
	recurrenceEnd := "never"
	recurrenceCount := "10"
	recurrenceUntil := ""

	// Create the form
	form := SimpleAddEventForm(&title, &eventType, &color, &calendarID, &startDate, &startTime, &duration, &endDate, &description, &location, &link, &formComplexity, &eventRecurrence, &recurrenceType, &recurrenceEnd, &recurrenceCount, &recurrenceUntil, calendars)

	// Show the form popup
	ShowFormPopup(form, "Add Event", func() tea.Cmd {
		// Build properties map from form data
		properties := make(map[string]string)

		// Always add the title
		properties["SUMMARY"] = title

		// Add optional properties
		if description != "" {
			properties["DESCRIPTION"] = description
		}
		if location != "" {
			properties["LOCATION"] = location
		}
		if link != "" {
			properties["URL"] = link
		}
		if color != "" {
			// Parse and normalize the color (e.g., 'b' -> '#0000FF')
			normalizedColor := util.ParseColor(color)
			if normalizedColor != "" {
				properties["X-COLOR"] = normalizedColor
			}
		}

		// Add recurrence rule if recurring
		if eventRecurrence == "recurring" {
			rrule, err := BuildRRuleFromFormData(&recurrenceType, &recurrenceEnd, &recurrenceCount, &recurrenceUntil, &startDate, weekLocation)
			if err == nil && rrule != "" {
				properties["RRULE"] = rrule
			}
		}

		// Parse times based on event type
		var startTimeParsed, endTimeParsed time.Time
		var isAllDay bool

		if eventType == "allday" {
			// All-day event
			isAllDay = true
			startDateParsed, err := time.ParseInLocation("2006-01-02", startDate, weekLocation)
			if err != nil {
				return func() tea.Msg { return types.VimErrorMsg{Error: fmt.Sprintf("Invalid start date: %v", err)} }
			}
			startTimeParsed = startDateParsed
			endTimeParsed = startDateParsed.AddDate(0, 0, 1) // Next day
		} else {
			// Timed event
			isAllDay = false
			if startDate == "" || startTime == "" {
				return func() tea.Msg { return types.VimErrorMsg{Error: "Start date and time are required for timed events"} }
			}

			dateTimeStr := startDate + " " + startTime
			var err error
			startTimeParsed, err = time.ParseInLocation("2006-01-02 15:04", dateTimeStr, weekLocation)
			if err != nil {
				return func() tea.Msg { return types.VimErrorMsg{Error: fmt.Sprintf("Invalid start date/time: %v", err)} }
			}

			// Parse duration
			durationMins := 60 // default
			if duration != "" {
				if dur, err := strconv.Atoi(duration); err == nil {
					durationMins = dur
				}
			}
			endTimeParsed = startTimeParsed.Add(time.Duration(durationMins) * time.Minute)
		}

		// Call the create handler
		return onCreate(properties, startTimeParsed, endTimeParsed, isAllDay, calendarID)
	}, nil)
}

// ShowEditEvent shows a comprehensive edit form for non-recurring events
//
// ARCHITECTURE: Unified editing system that branches by event type
//
// This function enables all-day event editing from the agenda picker, which was previously
// blocked due to complexity around iCal all-day event handling. The key challenges solved:
//
// 1. iCal all-day events use exclusive end dates (DTEND = last day + 1)
// 2. All-day events require VALUE=DATE parameters to remain visible
// 3. Users expect inclusive date ranges in forms
//
// The implementation branches to specialized handlers that manage these complexities
// while presenting a consistent editing experience for both timed and all-day events.
func ShowEditEvent(event *ical.Event, onUpdate EventUpdateHandler) {
	// Block recurring events only (all-day editing is now supported!)
	if util.IsRecurring(event) {
		return
	}

	// Branch to specialized handlers based on event type
	if util.IsAllDayEvent(event) {
		showEditAllDayEvent(event, onUpdate)    // Handles inclusive/exclusive date conversion
	} else {
		showEditTimedEvent(event, onUpdate)     // Handles timezone-safe time editing
	}
}

// showEditTimedEvent handles editing for timed events (existing logic)
func showEditTimedEvent(event *ical.Event, onUpdate EventUpdateHandler) {
	// Extract current values from event
	title := util.GetEventTitle(event)
	description := util.GetEventDescription(event)
	location := util.GetEventLocation(event)
	link := util.GetEventLink(event)
	color := util.GetEventColor(event, "")

	// Extract timing information
	startLocalTime, startErr := timezone.GetEventStartTime(event)
	endLocalTime, endErr := timezone.GetEventEndTime(event)
	if startErr != nil || endErr != nil {
		startLocalTime = timezone.NewLocalTime(time.Now())
		endLocalTime = timezone.NewLocalTime(time.Now().Add(time.Hour))
	}
	
	startTime := startLocalTime.Time
	endTime := endLocalTime.Time
	startDate := startTime.Format("2006-01-02")
	startTimeStr := startTime.Format("15:04")

	// Calculate duration in minutes
	duration := "60"
	if !startTime.IsZero() && !endTime.IsZero() {
		durationMins := int(endTime.Sub(startTime).Minutes())
		duration = fmt.Sprintf("%d", durationMins)
	}

	form := SimpleEditEventForm(&title, &description, &location, &link, &color, &startDate, &startTimeStr, &duration)

	ShowFormPopup(form, "Edit Event", func() tea.Cmd {
		properties := buildBaseProperties(title, description, location, link, color)
		addTimedEventProperties(properties, event, startDate, startTimeStr, duration)
		return onUpdate(properties)
	}, nil)
}

// showEditAllDayEvent handles editing for all-day events
func showEditAllDayEvent(event *ical.Event, onUpdate EventUpdateHandler) {
	// Extract current values from event
	title := util.GetEventTitle(event)
	description := util.GetEventDescription(event)
	location := util.GetEventLocation(event)
	link := util.GetEventLink(event)
	color := util.GetEventColor(event, "")

	// Extract date information for all-day events
	startLocalTime, startErr := timezone.GetEventStartTime(event)
	endLocalTime, endErr := timezone.GetEventEndTime(event)
	if startErr != nil || endErr != nil {
		startLocalTime = timezone.NewLocalTime(time.Now())
		endLocalTime = timezone.NewLocalTime(time.Now().AddDate(0, 0, 1))
	}
	
	startDate := startLocalTime.Time.Format("2006-01-02")
	
	// CRITICAL: iCal All-Day Event End Date Handling
	// 
	// Per RFC 5545, all-day events use EXCLUSIVE end dates:
	// - Single day event on Dec 21: DTSTART=20241221, DTEND=20241222
	// - Multi-day Dec 21-23: DTSTART=20241221, DTEND=20241224 (ends before Dec 24)
	//
	// But users expect INCLUSIVE dates in forms:
	// - Single day: "Start: Dec 21, End: Dec 21" 
	// - Multi-day: "Start: Dec 21, End: Dec 23"
	//
	// So we subtract 1 day from DTEND for display in the form
	displayEndDate := endLocalTime.Time.AddDate(0, 0, -1)
	endDate := displayEndDate.Format("2006-01-02")

	form := SimpleEditAllDayEventForm(&title, &description, &location, &link, &color, &startDate, &endDate)

	ShowFormPopup(form, "Edit All-Day Event", func() tea.Cmd {
		properties := buildBaseProperties(title, description, location, link, color)
		addAllDayEventProperties(properties, startDate, endDate)
		return onUpdate(properties)
	}, nil)
}

// buildBaseProperties creates the common properties map
func buildBaseProperties(title, description, location, link, color string) map[string]string {
	properties := map[string]string{
		"SUMMARY":       title,
		"LAST-MODIFIED": time.Now().UTC().Format("20060102T150405Z"),
	}

	if description != "" {
		properties["DESCRIPTION"] = description
	}
	if location != "" {
		properties["LOCATION"] = location
	}
	if link != "" {
		properties["URL"] = link
	}
	if color != "" {
		if normalizedColor := util.ParseColor(color); normalizedColor != "" {
			properties["X-COLOR"] = normalizedColor
		}
	}

	return properties
}

// addTimedEventProperties adds timing properties for timed events
func addTimedEventProperties(properties map[string]string, event *ical.Event, startDate, startTimeStr, duration string) {
	if startDate == "" || startTimeStr == "" {
		return
	}

	// Get original timezone context
	originalDtstart := event.Props.Get("DTSTART")
	originalDtstartValue := ""
	if originalDtstart != nil {
		originalDtstartValue = originalDtstart.Value
	}
	
	// Parse new start time
	icalStartValue, err := timezone.ConvertFormInputToICalFormat(startDate, startTimeStr, originalDtstartValue)
	if err != nil {
		return
	}

	// Calculate new end time from duration
	if durationMins, err := strconv.Atoi(duration); err == nil && durationMins > 0 {
		startLocalTime, parseErr := timezone.ParseFormInput(startDate, startTimeStr)
		if parseErr != nil {
			return
		}
		newEndTime := startLocalTime.Add(time.Duration(durationMins) * time.Minute)
		
		// Get original DTEND timezone context
		originalDtend := event.Props.Get("DTEND")
		originalDtendValue := ""
		if originalDtend != nil {
			originalDtendValue = originalDtend.Value
		}
		
		properties["DTSTART"] = icalStartValue
		properties["DTEND"] = timezone.PreserveTimezoneFormat(newEndTime, originalDtendValue)
	}
}

// addAllDayEventProperties adds date properties for all-day events
// 
// CRITICAL: Handles the iCal all-day event exclusive end date conversion
// 
// User form shows INCLUSIVE dates: "Start: Dec 21, End: Dec 21" (single day)
// But iCal requires EXCLUSIVE dates: DTSTART=20241221, DTEND=20241222
// 
// This function converts from user-friendly inclusive dates to iCal exclusive format
func addAllDayEventProperties(properties map[string]string, startDate, endDate string) {
	if startDate == "" || endDate == "" {
		return
	}

	// Parse the user-input dates (inclusive format)
	startDateParsed, startErr := time.Parse("2006-01-02", startDate)
	endDateParsed, endErr := time.Parse("2006-01-02", endDate)
	
	if startErr == nil && endErr == nil {
		// CRITICAL CONVERSION: Add 1 day to make DTEND exclusive
		// User input "End: Dec 21" becomes DTEND=20241222 (exclusive)
		// This ensures:
		// - Single day event: user enters same start/end → DTEND = start + 2 days
		// - Multi-day event: user enters "21st-23rd" → DTEND = 24th (ends before 24th)
		exclusiveEndDate := endDateParsed.AddDate(0, 0, 1)
		
		// Format as iCal DATE values (YYYYMMDD format, no time component)
		dtstart := startDateParsed.Format("20060102")
		dtend := exclusiveEndDate.Format("20060102")
		
		properties["DTSTART"] = dtstart
		properties["DTEND"] = dtend
	}
}
