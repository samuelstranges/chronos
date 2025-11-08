package navigation

import (
	"github.com/samuelstranges/chronos/ical_crud"
	"github.com/samuelstranges/chronos/types"
)

// Navigate is the single entry point for all navigation
// Returns true if navigation succeeded, false if no match found
func Navigate(weekModel *types.WeekModel, eventManager *ical_crud.EventManager, direction Direction, criteria NavigationCriteria) bool {
	// Step 1: Try grid-based navigation within current week
	if navigateWithinWeek(weekModel, direction, criteria) {
		return true
	}

	// Step 2: Fall back to time-based cross-week navigation
	if navigateAcrossWeeks(weekModel, eventManager, direction, criteria) {
		return true
	}

	return false
}