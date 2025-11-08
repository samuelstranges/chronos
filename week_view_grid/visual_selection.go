package week_view_grid

import (
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// GetEventsInVisualSelectionByTime finds all events within the visual selection time range
// across all visible calendars, supporting multi-week selections
func GetEventsInVisualSelectionByTime(
	weekModel *types.WeekModel,
	eventManager types.EventManagerInterface,
) []*ical.Event {
	if !weekModel.IsVisualMode {
		return nil
	}

	// Calculate time range for visual selection using cell boundaries
	anchorTime := util.CellToTime(weekModel.VisualAnchor.Cell, weekModel.CurrentZoom, weekModel.VisualAnchorWeek, weekModel.VisualAnchor.Day)
	cursorTime := util.CellToTime(weekModel.Cursor.Cell, weekModel.CurrentZoom, weekModel.CurrentlyViewedWeek, weekModel.Cursor.Day)

	// Determine start and end times using cell boundaries
	// Start at beginning of earlier cell, end at end of later cell
	var startTime, endTime time.Time
	cellDuration := time.Duration(weekModel.CurrentZoom) * time.Minute

	if cursorTime.Before(anchorTime) {
		startTime = cursorTime                 // Start of cursor cell
		endTime = anchorTime.Add(cellDuration) // End of anchor cell
	} else {
		startTime = anchorTime                 // Start of anchor cell
		endTime = cursorTime.Add(cellDuration) // End of cursor cell
	}

	// Get all visible calendars
	calendars := eventManager.GetCalendarsForDisplay()
	var selectedEvents []*ical.Event

	// Search through all events in visible calendars
	for _, calendar := range calendars {
		for _, component := range calendar.Children {
			if component.Name != "VEVENT" {
				continue
			}

			// Get event start and end times
			eventStartProp := component.Props.Get("DTSTART")
			eventEndProp := component.Props.Get("DTEND")
			if eventStartProp == nil || eventEndProp == nil {
				continue
			}

			// Use timezone-safe methods to get event times
			eventWrapper := &ical.Event{Component: component}
			eventStartTime, startErr := timezone.GetEventStartTime(eventWrapper)
			eventEndTime, endErr := timezone.GetEventEndTime(eventWrapper)
			if startErr != nil || endErr != nil {
				continue
			}
			
			// Convert to time.Time for compatibility
			eventStart := eventStartTime.Time
			eventEnd := eventEndTime.Time

			// Skip all-day events - they should be handled separately in the all-day events layer
			// Skip recurring events in visual mode for safety - prevents accidental deletion of entire series
			event := &ical.Event{Component: component}
			if util.IsAllDayEvent(event) || util.IsRecurring(event) {
				continue
			}

			// Check if event overlaps with selection time range
			if eventOverlapsTimeRange(eventStart, eventEnd, startTime, endTime) {
				selectedEvents = append(selectedEvents, event)
			}
		}
	}

	return selectedEvents
}

// eventOverlapsTimeRange checks if an event time range overlaps with the selection time range
func eventOverlapsTimeRange(eventStart, eventEnd, selectionStart, selectionEnd time.Time) bool {
	// Events overlap if event starts before selection ends (exclusive), and event ends after selection starts (exclusive)
	// This ensures proper cell boundary handling - events at exact boundaries belong to adjacent cells
	return util.TimeRangesOverlapExclusive(eventStart, eventEnd, selectionStart, selectionEnd)
}
