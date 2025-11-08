package util

import (
	"time"

	"github.com/samuelstranges/chronos/types"
)

// TimeRangesOverlapExclusive checks if two time ranges overlap with exclusive end times
// This is useful for events where DTEND is non-inclusive (RFC 5545 standard)
func TimeRangesOverlapExclusive(start1, end1, start2, end2 time.Time) bool {
	// Two ranges overlap if: start1 < end2 && start2 < end1
	// Exclusive version - touching boundaries don't count as overlap
	return start1.Before(end2) && start2.Before(end1)
}

// EventInstanceOverlapsTimeRange checks if an EventInstance overlaps with a specific time range
func EventInstanceOverlapsTimeRange(instance *types.EventInstance, rangeStart, rangeEnd time.Time) bool {
	if instance == nil {
		return false
	}
	return TimeRangesOverlapExclusive(instance.ComputedStart.Time, instance.ComputedEnd.Time, rangeStart, rangeEnd)
}

// CellRangesOverlap checks if two cell ranges overlap within a day
// This is the core collision detection for visual grid rendering
func CellRangesOverlap(startCellA, endCellA, startCellB, endCellB int) bool {
	// Cell ranges overlap if: startA <= endB && startB <= endA
	return startCellA <= endCellB && startCellB <= endCellA
}

// TimeIsWithinRange checks if a time falls within a given time range (inclusive)
func TimeIsWithinRange(t, minTime, maxTime time.Time) bool {
	return !t.Before(minTime) && !t.After(maxTime)
}