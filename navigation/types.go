package navigation

import "time"

// Direction represents navigation direction
type Direction int

const (
	Forward Direction = iota
	Backward
)

// Position represents a cursor position in the calendar
type Position struct {
	Week        time.Time  // Week start
	Day         int        // 0-6 (Sunday-Saturday)
	Cell        int        // Time slot within day
	EventColumn int        // Column within collision group
}