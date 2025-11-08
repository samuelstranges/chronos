package ical_crud

// IMPORTANT: ALL EventManager methods that modify calendar data (add/remove/edit events,
// add/remove calendars, etc.) MUST accept a *types.WeekModel parameter and call
// autoRefresh(weekModel) to update the display grid.
//
// This ensures that:
// 1. The visual display is immediately updated after data changes
// 2. All event grids are rebuilt with the new calendar state
// 3. Users see changes instantly without manual refresh
// 4. The architecture remains consistent across all operations
//
// Methods like AddCalendar, RemoveCalendar, CreateEvent, DeleteEvent, etc. follow this pattern.
// DO NOT create calendar/event modification methods without weekModel + autoRefresh!

import (
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/storage"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// CopiedEvent represents a copied event with its source calendar
type CopiedEvent struct {
	Event    *ical.Event
	Calendar *ical.Calendar
}

type EventManager struct {
	calendars         []*ical.Calendar                      // EventManager owns the calendar data
	calendarMap       map[string]*ical.Calendar             // ID -> calendar reference for form integration
	calendarVisible   map[string]bool                       // ID -> visibility state for calendar filtering
	changeTracker     *util.ChangeTracker[EventsChanged]     // Generic change tracker for undo/redo
	copiedEvents      []CopiedEvent                         // Stores multiple copied events for paste operations
	storage           storage.CalendarStorage               // Storage interface for persistence operations
	skipServerSync    bool                                  // Flag to skip server sync (for undo/redo rollback)
}

// New creates an EventManager with calendars from storage (stable IDs preserved)
func New(calendarMap map[string]*ical.Calendar, storageImpl storage.CalendarStorage) *EventManager {
	// Convert map to slice for backward compatibility with existing code
	calendars := make([]*ical.Calendar, 0, len(calendarMap))
	for _, cal := range calendarMap {
		calendars = append(calendars, cal)
	}
	
	em := &EventManager{
		calendars:       calendars,     // Keep slice for compatibility
		calendarMap:     calendarMap,   // Use stable IDs directly from storage!
		calendarVisible: make(map[string]bool),
		changeTracker:   util.NewChangeTracker[EventsChanged](types.MaxUndoOperations),
		storage:         storageImpl,
	}

	// Set all calendars visible by default (no ID regeneration needed!)
	for calendarID := range calendarMap {
		em.calendarVisible[calendarID] = true
	}

	return em
}

// REMOVED: autoRefresh() and SetRefreshCallback() - replaced with tea.Cmd pattern
// All operations now return (error, tea.Cmd) for proper Bubble Tea architecture

// ReloadFromStorage reloads all calendars from storage (for background sync)
func (em *EventManager) ReloadFromStorage() error {
	calendarMap, err := em.storage.LoadCalendars()
	if err != nil {
		return err
	}

	// Replace the calendar map with freshly loaded data
	em.calendarMap = calendarMap

	// Rebuild calendars slice for backward compatibility
	em.calendars = make([]*ical.Calendar, 0, len(calendarMap))
	for _, cal := range calendarMap {
		em.calendars = append(em.calendars, cal)
	}

	return nil
}
