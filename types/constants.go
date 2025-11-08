// Package types provides core type definitions, constants, and structures used throughout the Chronos calendar application.
package types

import (
	"fmt"
	"time"
)

const (
	// TimeColumnBorder is the width of the right border for the time column
	TimeColumnBorder = 1
	// DayColumnBorders is the total width of borders for all day columns
	DayColumnBorders = 7
	// TotalBorders is the combined width of all column borders
	TotalBorders     = TimeColumnBorder + DayColumnBorders

	// MinVisibleCellsPerDay is the minimum number of time cells visible per day
	MinVisibleCellsPerDay = 1

	// CellPaddingLeft is the left padding for calendar cells
	CellPaddingLeft        = 0
	// CellPaddingRight is the right padding for calendar cells
	CellPaddingRight       = 0
	// TimeColumnLeftPadding is the left padding for the time column
	TimeColumnLeftPadding  = 1
	// TimeColumnRightPadding is the right padding for the time column
	TimeColumnRightPadding = 1

	// Calendar time types
	DaysPerWeek = 7
	MaxDayIndex = 6 // 0-indexed days (0=Sunday, 6=Saturday)
	MinsPerHour = 60
	HoursPerDay = 24
	MinsPerDay  = MinsPerHour * HoursPerDay

	// UI Reserved Lines
	// TODO: build these using lipgloss get height methods?
	AllDayCells      = 3
	AllDayCountRow   = 1 // New count row showing "X all-day events"
	AllDaySeperator  = 1 //
	MonthHeaderLines = 1
	DayTitleLines    = 1

	HeaderLines = MonthHeaderLines + DayTitleLines + AllDayCountRow + AllDayCells
	FooterLines = 1

	// Color types
	HeaderBackgroundColor = "240" // Background color for headers and count rows
	DefaultEventLength    = 60    // Default event duration in minutes

	TimeFormat = "%02d:%02d" // displayed in week mode for example

	// UI Colors
	CurrentTimeRowColor       = "#e0af68"
	TimeCellCursorRowColor    = "#9ece6a"
	BrightenBackgroundPercent = 270
	BrightenEventsPercent     = 80
	EventTextBlackColor       = "#000000" // Black color for toggled event text

	DefaultEventColor = "#FF2968" // Fallback color when no calendar or event color specified

	// Time duration types
	ErrorDisplayDuration = 5 * time.Second // How long to show error messages
	ErrorDisplaySeconds  = 5               // How long to show error messages (in seconds)

	// UI interaction types
	KeyHintTimeoutSeconds = 3 // Seconds before showing key hints
	TwoDigitLength        = 2 // Expected length for two-digit inputs
	FourDigitLength       = 4 // Expected length for four-digit inputs

	// Undo system
	MaxUndoOperations = 50

	// Color manipulation types
	ColorDarkenPercent = 0.3   // For darkening colors
	LightnessScale     = 100.0 // Scale for lightness calculations

	// Layout and positioning types
	DefaultMarginTop    = 2  // Default top margin
	DefaultMarginBottom = 2  // Default bottom margin
	HeaderSpacing       = 3  // Spacing in headers
	FormSpacing         = 3  // Spacing in forms
	MinimumWidth        = 10 // Minimum widget width
	StandardWidth       = 20 // Standard widget width
	WideWidth           = 30 // Wide widget width
	ExtraWideWidth      = 40 // Extra wide widget width
	VeryWideWidth       = 45 // Very wide widget width
	MaxWidth            = 50 // Maximum widget width
	LargeWidth          = 55 // Large widget width
	ExtendedWidth       = 60 // Extended widget width

	// Alpha/transparency types
	ReducedOpacity = 0.8 // Reduced opacity for visual effects
	CellOpacity    = 1.8 // Cell opacity multiplier

	// Hash types for event manager
	HashShift16  = 16   // Bit shift for 16-bit operations
	HashMask0x40 = 0x40 // Hash mask for 0x40
	HashMask0x80 = 0x80 // Hash mask for 0x80

	// Picker types
	PickerMinItems = 6 // Minimum items to show picker

	// Time interval types
	DefaultMinutesInterval = 30 // Default interval in minutes
	SecondsPerMinute       = 60 // Seconds in a minute
	HalfDivisor            = 2  // Divisor for half calculations

	// Form validation types
	WeekdayIntervalDivisor = 7 // Divisor for weekday intervals

	// Return codes for form operations
	FormReturnCodeTwo   = 2 // Return code 2
	FormReturnCodeThree = 3 // Return code 3
	FormReturnCodeFour  = 4 // Return code 4

	// Calendar display types
	HourDisplayDivisor = 24 // Divisor for hour display calculations
	MaxRetryAttempts   = 5  // Maximum retry attempts for operations
)

// Const but a var
var (
	TimeColumnAndPaddingWidth          = len(fmt.Sprintf(TimeFormat, 0, 0)) + TimeColumnLeftPadding + TimeColumnRightPadding
	TimeColumnAndPaddingAndBorderWidth = TimeColumnAndPaddingWidth + 1
)

// Calendar component and property types (external protocol values)
const (
	// ICalComponentEvent represents the VEVENT component type in iCal
	ICalComponentEvent = "VEVENT"

	// ICalValueTypeDate represents the DATE value type in iCal
	ICalValueTypeDate = "DATE"

	// UnknownCalendar is the default name for calendars without a name
	UnknownCalendar = "Unknown Calendar"
)

// EventType represents the type of event
type EventType string

const (
	// EventTypeAllDay represents an all-day event type
	EventTypeAllDay EventType = "allday"
	// EventTypeNormal represents a normal timed event
	EventTypeNormal EventType = "normal"
)

// EventRecurrence represents whether an event is recurring
type EventRecurrence string

const (
	// EventRecurrenceRecurring represents a recurring event
	EventRecurrenceRecurring EventRecurrence = "recurring"
	// EventRecurrenceOnce represents a one-time event
	EventRecurrenceOnce EventRecurrence = "once"
)

// RecurrencePattern represents the pattern of recurrence
type RecurrencePattern string

const (
	// RecurrencePatternWeekday represents weekday recurrence
	RecurrencePatternWeekday RecurrencePattern = "weekday"

	// RecurrencePatternMonthly represents monthly recurrence
	RecurrencePatternMonthly RecurrencePattern = "monthly"

	// RecurrencePatternYearly represents yearly recurrence
	RecurrencePatternYearly RecurrencePattern = "yearly"
)

// RecurrenceEndType represents how a recurrence ends
type RecurrenceEndType string

const (
	// RecurrenceEndCount represents ending by count
	RecurrenceEndCount RecurrenceEndType = "count"

	// RecurrenceEndUntil represents ending by date
	RecurrenceEndUntil RecurrenceEndType = "until"
)

// UI Interaction strings
const (
	// Key names for UI navigation
	KeyDown  = "down"
	KeyEnter = "enter"
	KeyEsc   = "esc"

	// Default event title
	UntitledEvent = "Untitled Event"

	// Color constants
	ColorBlack = "#000000"

	// Time search constants
	DaysInYear = 365

	// UI popup width constants
	PopupWidthLarge = 60

	// UI padding constants  
	PaddingHorizontalStandard = 2
	PaddingVerticalStandard   = 1
)
