package util

import (
	"time"
)

// ChangeType represents the type of change made
type ChangeType int

const (
	ChangeTypeAdd ChangeType = iota
	ChangeTypeDelete
	ChangeTypeEdit
)

// Change represents a single change that can be undone/redone
// This is generic and can hold any type of data for any domain
type Change[T any] struct {
	Type      ChangeType
	Data      T         // The data that was added/deleted/modified
	Context   T         // Additional context (e.g., which container it belongs to)
	Timestamp time.Time // When the change was made

	// For edit changes, store the before state
	PreviousData T // Only used for edits
}

// BatchChange represents a group of changes that should be undone/redone together
type BatchChange[T any] struct {
	Changes     []Change[T]
	Description string // Description of the batch operation (e.g., "Paste 3 events")
	Timestamp   time.Time
}

// ChangeTracker manages undo/redo history for any type of data
// All operations are treated as batches internally for consistency
type ChangeTracker[T any] struct {
	undoStack    []BatchChange[T]
	redoStack    []BatchChange[T]
	maxHistory   int
	currentBatch []Change[T] // Accumulates changes for current batch operation
}


// getChangeTypeString converts ChangeType to string for logging
func getChangeTypeString(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeAdd:
		return "ADD"
	case ChangeTypeDelete:
		return "DELETE"
	case ChangeTypeEdit:
		return "EDIT"
	default:
		return "UNKNOWN"
	}
}

// NewChangeTracker creates a new generic change tracker
func NewChangeTracker[T any](maxHistory int) *ChangeTracker[T] {
	if maxHistory <= 0 {
		maxHistory = 50 // Default max history
	}

	return &ChangeTracker[T]{
		undoStack:    make([]BatchChange[T], 0, maxHistory),
		redoStack:    make([]BatchChange[T], 0, maxHistory),
		maxHistory:   maxHistory,
		currentBatch: make([]Change[T], 0),
	}
}

// RecordChange adds a change to the current batch
func (ct *ChangeTracker[T]) RecordChange(change Change[T]) {
	change.Timestamp = time.Now()
	ct.currentBatch = append(ct.currentBatch, change)
	
}

// CommitBatch completes the current batch and commits all changes as a single undo unit
func (ct *ChangeTracker[T]) CommitBatch(description string) {
	if len(ct.currentBatch) == 0 {
		return // Nothing to commit
	}

	batchChange := BatchChange[T]{
		Changes:     append([]Change[T](nil), ct.currentBatch...), // Copy the changes
		Description: description,
		Timestamp:   time.Now(),
	}

	// Add to undo stack
	ct.undoStack = append(ct.undoStack, batchChange)

	// Trim if we exceed max history
	if len(ct.undoStack) > ct.maxHistory {
		ct.undoStack = ct.undoStack[1:]
	}

	// Clear redo stack when new change is made
	ct.redoStack = ct.redoStack[:0]

	// Clear current batch
	ct.currentBatch = ct.currentBatch[:0]
	
}

// CanUndo returns true if there are changes to undo
func (ct *ChangeTracker[T]) CanUndo() bool {
	return len(ct.undoStack) > 0
}

// CanRedo returns true if there are changes to redo
func (ct *ChangeTracker[T]) CanRedo() bool {
	return len(ct.redoStack) > 0
}

// Undo removes and returns the most recent batch change from undo stack
func (ct *ChangeTracker[T]) Undo() *BatchChange[T] {
	if !ct.CanUndo() {
		return nil
	}

	// Get the batch change; remove from undo stack; add to redo stack
	batchChange := ct.undoStack[len(ct.undoStack)-1]
	ct.undoStack = ct.undoStack[:len(ct.undoStack)-1]
	ct.redoStack = append(ct.redoStack, batchChange)

	return &batchChange
}

// Redo removes and returns the most recent batch change from redo stack
func (ct *ChangeTracker[T]) Redo() *BatchChange[T] {
	if !ct.CanRedo() {
		return nil
	}

	// Get the batch change; remove from redo stack and add to undo stack
	batchChange := ct.redoStack[len(ct.redoStack)-1]
	ct.redoStack = ct.redoStack[:len(ct.redoStack)-1]
	ct.undoStack = append(ct.undoStack, batchChange)

	return &batchChange
}