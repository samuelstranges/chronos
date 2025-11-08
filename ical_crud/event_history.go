package ical_crud

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// SingleEventChange represents a single event change
type SingleEventChange struct {
	EventData         *ical.Event    // The event that was added/deleted/modified (after state)
	PreviousEventData *ical.Event    // Before state (for edits)
	Calendar          *ical.Calendar // Which calendar the event belongs to
}

// EventsChanged represents one or more event changes that should be undone/redone together
type EventsChanged struct {
	Changes []SingleEventChange // Array of individual event changes
}

// Undo reverses the last batch of changes
func (em *EventManager) Undo() (error, tea.Cmd) {
	batchChange := em.changeTracker.Undo()
	if batchChange == nil {
		return fmt.Errorf("nothing to undo"), nil
	}

	// Set skipServerSync flag to prevent recordChange from syncing during rollback
	em.skipServerSync = true

	// Process changes in reverse order for undo
	for i := len(batchChange.Changes) - 1; i >= 0; i-- {
		change := batchChange.Changes[i]
		eventsChanged := change.Data
		// Process each individual event change in reverse order
		for j := len(eventsChanged.Changes) - 1; j >= 0; j-- {
			singleChange := eventsChanged.Changes[j]
			em.applyUndoChange(change.Type, singleChange)
		}
	}

	// Clear the flag
	em.skipServerSync = false

	// Sync to storage using granular operations (inverse operations for undo)
	return nil, em.syncUndoRedoToStorage(batchChange, true)
}

// Redo reapplies the last undone batch of changes
func (em *EventManager) Redo() (error, tea.Cmd) {
	batchChange := em.changeTracker.Redo()
	if batchChange == nil {
		return fmt.Errorf("nothing to redo"), nil
	}

	// Set skipServerSync flag to prevent recordChange from syncing during rollback
	em.skipServerSync = true

	// Process changes in forward order for redo
	for _, change := range batchChange.Changes {
		eventsChanged := change.Data
		// Process each individual event change
		for _, singleChange := range eventsChanged.Changes {
			em.applyRedoChange(change.Type, singleChange)
		}
	}

	// Clear the flag
	em.skipServerSync = false

	// Sync to storage using granular operations (original operations for redo)
	return nil, em.syncUndoRedoToStorage(batchChange, false)
}

// applyUndoChange applies a single undo operation
func (em *EventManager) applyUndoChange(changeType util.ChangeType, singleChange SingleEventChange) {
	switch changeType {
	case util.ChangeTypeAdd:
		// Undo add: remove the event
		eventUID := util.GetEventUID(singleChange.EventData)
		_, index := util.FindEventByUIDInCalendar(singleChange.Calendar, eventUID)
		if index != -1 {
			util.RemoveEventFromCalendar(singleChange.Calendar, index)
		}
	case util.ChangeTypeDelete:
		// Undo delete: add the event back
		singleChange.Calendar.Children = append(singleChange.Calendar.Children, singleChange.EventData.Component)
	case util.ChangeTypeEdit:
		// Undo edit: restore the original event state using PreviousEventData
		if singleChange.PreviousEventData != nil {
			targetUID := util.GetEventUID(singleChange.PreviousEventData)
			em.replaceEventComponentByUID(targetUID, singleChange.PreviousEventData.Component)
		}
	}
}

// applyRedoChange applies a single redo operation
func (em *EventManager) applyRedoChange(changeType util.ChangeType, singleChange SingleEventChange) {
	switch changeType {
	case util.ChangeTypeAdd:
		// Redo add: add the event back
		singleChange.Calendar.Children = append(singleChange.Calendar.Children, singleChange.EventData.Component)
	case util.ChangeTypeDelete:
		// Redo delete: remove the event again
		eventUID := util.GetEventUID(singleChange.EventData)
		_, index := util.FindEventByUIDInCalendar(singleChange.Calendar, eventUID)
		if index != -1 {
			util.RemoveEventFromCalendar(singleChange.Calendar, index)
		}
	case util.ChangeTypeEdit:
		// Redo edit: restore the modified event state using EventData
		if singleChange.EventData != nil {
			targetUID := util.GetEventUID(singleChange.EventData)
			em.replaceEventComponentByUID(targetUID, singleChange.EventData.Component)
		}
	}
}

// CanUndo returns true if there are changes to undo
func (em *EventManager) CanUndo() bool {
	return em.changeTracker.CanUndo()
}

// CanRedo returns true if there are changes to redo
func (em *EventManager) CanRedo() bool {
	return em.changeTracker.CanRedo()
}

// recordChange records event changes and saves to storage
// This is the ONLY method that should be used for change recording - it handles all the boilerplate
// For single changes: EventsChanged{Changes: []SingleEventChange{{...}}}
// Returns a tea.Cmd for async server sync (CalDAV only) and UI refresh
func (em *EventManager) recordChange(changeType util.ChangeType, eventsChanged EventsChanged, description string) tea.Cmd {
	change := util.Change[EventsChanged]{
		Type: changeType,
		Data: eventsChanged,
	}
	em.changeTracker.RecordChange(change)
	em.changeTracker.CommitBatch(description)

	// Skip server sync if this is a rollback operation (undo/redo)
	if em.skipServerSync {
		return func() tea.Msg {
			return types.RefreshMsg{}
		}
	}

	// Check if we're using CalDAV (async) or file storage (sync)
	isCalDAV := em.storage.GetStorageType() == "caldav"

	// For CalDAV, use async sync
	if isCalDAV {
		return tea.Batch(
			em.SyncToServerAsync(changeType, eventsChanged),
			func() tea.Msg {
				return types.RefreshMsg{}
			},
		)
	}

	// For file storage, save all calendars synchronously
	if err := em.storage.SaveCalendars(em.calendarMap); err != nil {
		return func() tea.Msg {
			return types.VimErrorMsg{Error: fmt.Sprintf("Failed to save: %v", err)}
		}
	}

	// Return refresh command
	return func() tea.Msg {
		return types.RefreshMsg{}
	}
}

// findCalendarID finds the calendar ID for a given calendar object
func (em *EventManager) findCalendarID(calendar *ical.Calendar) string {
	for calendarID, cal := range em.calendarMap {
		if cal == calendar {
			return calendarID
		}
	}
	return ""
}

// SyncToServerAsync performs async CalDAV sync with undo on failure
// This allows instant local updates while syncing to server in background
// If sync fails, automatically undoes the local change
func (em *EventManager) SyncToServerAsync(changeType util.ChangeType, eventsChanged EventsChanged) tea.Cmd {
	// Only for CalDAV storage
	if em.storage.GetStorageType() != "caldav" {
		return nil
	}

	return func() tea.Msg {
		// Perform storage operations in background
		for _, singleChange := range eventsChanged.Changes {
			if singleChange.EventData == nil {
				continue
			}

			eventUID := singleChange.EventData.Props.Get("UID")
			if eventUID == nil {
				continue
			}

			calendarID := em.findCalendarID(singleChange.Calendar)
			if calendarID == "" {
				continue
			}

			var err error
			switch changeType {
			case util.ChangeTypeDelete:
				err = em.storage.DeleteEvent(calendarID, eventUID.Value)
			case util.ChangeTypeAdd, util.ChangeTypeEdit:
				_, err = em.storage.SaveEvent(calendarID, singleChange.EventData)
			}

			if err != nil {
				// Rollback using undo system (with skipServerSync flag to prevent loop)
				em.skipServerSync = true
				em.Undo()
				em.skipServerSync = false
				return types.VimErrorMsg{Error: fmt.Sprintf("Server sync failed, changes undone: %v", err)}
			}
		}
		return nil // Success - no message needed
	}
}

// getStorageOperation determines the correct storage operation for undo/redo
// Returns the event to use and whether it should be deleted
func getStorageOperation(changeType util.ChangeType, singleChange SingleEventChange, isUndo bool) (*ical.Event, bool) {
	if isUndo {
		// UNDO: Apply inverse of the original operation
		switch changeType {
		case util.ChangeTypeDelete:
			// Undoing delete → event was added back → save it
			return singleChange.EventData, false
		case util.ChangeTypeAdd:
			// Undoing add → event was removed → delete it
			return singleChange.EventData, true
		case util.ChangeTypeEdit:
			// Undoing edit → previous state restored → save previous
			return singleChange.PreviousEventData, false
		}
	} else {
		// REDO: Apply the original operation again
		switch changeType {
		case util.ChangeTypeDelete:
			// Redoing delete → event was removed → delete it
			return singleChange.EventData, true
		case util.ChangeTypeAdd:
			// Redoing add → event was added back → save it
			return singleChange.EventData, false
		case util.ChangeTypeEdit:
			// Redoing edit → modified state restored → save modified
			return singleChange.EventData, false
		}
	}
	return nil, false
}

// syncUndoRedoToStorage syncs undo/redo changes to storage
func (em *EventManager) syncUndoRedoToStorage(batchChange *util.BatchChange[EventsChanged], isUndo bool) tea.Cmd {
	// Check if we're using CalDAV (async) or file storage (sync)
	isCalDAV := em.storage.GetStorageType() == "caldav"

	if isCalDAV {
		// For CalDAV, queue async sync commands
		var cmds []tea.Cmd
		for _, change := range batchChange.Changes {
			cmds = append(cmds, em.syncUndoRedoToServerAsync(change.Type, change.Data, isUndo))
		}
		cmds = append(cmds, func() tea.Msg { return types.RefreshMsg{} })
		return tea.Batch(cmds...)
	}

	// For file storage, save all calendars synchronously
	if err := em.storage.SaveCalendars(em.calendarMap); err != nil {
		return func() tea.Msg {
			return types.VimErrorMsg{Error: fmt.Sprintf("Failed to save undo/redo: %v", err)}
		}
	}

	return func() tea.Msg { return types.RefreshMsg{} }
}

// syncUndoRedoToServerAsync performs async CalDAV sync for undo/redo operations
// Unlike SyncToServerAsync, this doesn't attempt to undo on failure (since we're already in undo/redo)
func (em *EventManager) syncUndoRedoToServerAsync(changeType util.ChangeType, eventsChanged EventsChanged, isUndo bool) tea.Cmd {
	if em.storage.GetStorageType() != "caldav" {
		return nil
	}

	return func() tea.Msg {
		for _, singleChange := range eventsChanged.Changes {
			eventToUse, shouldDelete := getStorageOperation(changeType, singleChange, isUndo)
			if eventToUse == nil {
				continue
			}

			eventUID := eventToUse.Props.Get("UID")
			if eventUID == nil {
				continue
			}

			calendarID := em.findCalendarID(singleChange.Calendar)
			if calendarID == "" {
				continue
			}

			// Perform inverse operation
			var err error
			if shouldDelete {
				err = em.storage.DeleteEvent(calendarID, eventUID.Value)
			} else {
				_, err = em.storage.SaveEvent(calendarID, eventToUse)
			}

			if err != nil {
				// For undo/redo, just show error - don't try to undo since we're already in undo/redo
				return types.VimErrorMsg{Error: fmt.Sprintf("Server sync failed for undo/redo: %v", err)}
			}
		}
		return nil
	}
}