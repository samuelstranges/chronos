package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
)

// CalendarInfo holds calendar information for forms
type CalendarInfo struct {
	ID    string
	Name  string
	Color string
}

// ZoomLevel represents different time granularities for the calendar view
type ZoomLevel int

const (
	// Zoom30Min represents 30-minute time slots
	Zoom30Min ZoomLevel = 30
	// Zoom15Min represents 15-minute time slots
	Zoom15Min ZoomLevel = 15
	// Zoom5Min represents 5-minute time slots
	Zoom5Min ZoomLevel = 5
	// Zoom1Min represents 1-minute time slots
	Zoom1Min ZoomLevel = 1
)

// ZoomLevels defines the canonical ordering of zoom levels for grid arrays
var ZoomLevels = [4]ZoomLevel{Zoom30Min, Zoom15Min, Zoom5Min, Zoom1Min}

// VimMode represents different vim modes with associated display information
type VimMode int

const (
	// ModeNormal represents normal navigation mode
	ModeNormal VimMode = iota
	// ModeVisual represents visual selection mode
	ModeVisual
	// ModeSearch represents search/filter mode
	ModeSearch
)

// ModeInfo contains display information for vim modes
type ModeInfo struct {
	Name            string
	BackgroundColor string
	ForegroundColor string
}

// Mode definitions with their associated colors (lualine-style)
var ModeInfoMap = map[VimMode]ModeInfo{
	ModeNormal: {
		Name:            "NORMAL",
		BackgroundColor: "#98be65", // green
		ForegroundColor: "#ffffff", // white
	},
	ModeVisual: {
		Name:            "VISUAL",
		BackgroundColor: "#c678dd", // magenta/purple
		ForegroundColor: "#ffffff", // white
	},
	ModeSearch: {
		Name:            "SEARCH",
		BackgroundColor: "#e06c75", // red
		ForegroundColor: "#ffffff", // white
	},
}

// GetModeInfo returns the mode info for a given vim mode
func (mode VimMode) GetModeInfo() ModeInfo {
	if info, exists := ModeInfoMap[mode]; exists {
		return info
	}
	// Default to NORMAL if mode not found
	return ModeInfoMap[ModeNormal]
}

// GetModeStyle returns a lipgloss style for the given mode
func (mode VimMode) GetModeStyle() lipgloss.Style {
	info := mode.GetModeInfo()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(info.BackgroundColor)).
		Foreground(lipgloss.Color(info.ForegroundColor)).
		Bold(true)
}

type CursorPositionInWeek struct {
	Day         int // 0 to 6 (Sunday to Saturday)
	Cell        int // cell number based on zoom level
	EventColumn int // which event column within the cell (0 = leftmost)
}

type WeekModel struct {
	CurrentlyViewedWeek           time.Time
	Cursor                        CursorPositionInWeek
	Width                         int
	Height                        int
	CurrentZoom                   ZoomLevel
	CachedCellWidth               int
	CachedContentWidth            int
	CachedBackgroundColor         string
	CachedBackgroundSelectedColor string

	WeekEventGrids    [4]WeekEventGrid // one for each zoomlevel
	AllDayEventGrid   AllDayEventGrid  // all-day events with row assignments
	AllDayEventCounts []int            // accurate count of all-day events per day

	// Visual mode selection
	IsVisualMode     bool
	VisualAnchor     CursorPositionInWeek
	VisualAnchorWeek time.Time // Week the visual anchor belongs to

	// Search functionality
	SearchActive      bool
	SearchInput       string
	SearchLocked      bool   // true when search is "locked in" after Enter
	LockedSearchInput string // the locked search input
	CurrentSearchIdx  int    // current position in search results (-1 = no position)

	// Error display
	ErrorMessage string
	ErrorExpiry  time.Time

	// All-day events layer
	ShowAllDayEventsLayer   bool
	AllDayEventsTargetDay   int         // 0-6 (Sunday to Saturday)
	AllDayEventsPickerIndex int         // currently selected event index in the picker (automatically in picker mode when overlay is open)
	SelectedAllDayEvent     *ical.Event // temporarily stores selected all-day event for form operations

	// Agenda layer
	ShowAgendaLayer   bool
	AgendaTargetDay   int         // 0-6 (Sunday to Saturday)
	AgendaPickerIndex int         // currently selected event index in the picker (automatically in picker mode when overlay is open)
	SelectedEvent     *ical.Event // temporarily stores selected event for form operations

	// Recurring event instance tracking for picker operations
	RecurringInstanceDate time.Time // stores the specific instance date when operating on recurring events from pickers

	// All-day grid visibility
	ShowAllDayGrid bool // controls whether the all-day event grid section is visible

	// Event text color toggle
	EventTextBlack bool // Toggle for event text color (default: true, black: true)
}

// VimCommandBinding represents a vim command binding for keyhints
type VimCommandBinding struct {
	Key         string
	Description string
	Modes       []VimMode
}

// VimCommandFolder represents a vim command folder for keyhints
type VimCommandFolder struct {
	Prefix      string
	Description string
}

// VimActionMsg is sent when a vim command needs to trigger an action
type VimActionMsg struct {
	Action   string
	Argument int // Could be repeat count (5j) or numeric parameter (gh14, gt1430)
}

// ShowRecurringEditFormMsg is sent when recurring dialog completes and needs to show edit form
type ShowRecurringEditFormMsg struct {
	Scope RecurringEventOperationScope
}

// RecurringDialogChoiceMsg is sent when user makes a choice in the recurring dialog
type RecurringDialogChoiceMsg struct {
	Event         *ical.Event
	Scope         RecurringEventOperationScope // This/Future/All
	OperationType RecurringEventOperationType  // Edit/Delete
	Cancelled     bool                         // True if user cancelled
}

// VimErrorMsg is sent when a vim command encounters an error
type VimErrorMsg struct {
	Error string
}

// RefreshMsg is sent when the display needs to be refreshed
type RefreshMsg struct{}

// BackgroundSyncMsg is sent periodically to sync calendars from CalDAV server
type BackgroundSyncMsg struct{}

// VimStateInterface defines the interface for vim state operations to avoid circular imports
type VimStateInterface interface {
	SetMode(mode VimMode)
	GetMode() VimMode
	SetVisualAnchor(cursor CursorPositionInWeek)
	GetStatusText() string
}

// EventInstance represents a unified view of either a regular event or a recurring event instance
// This abstraction provides computed start/end times while preserving the original iCal event
type EventInstance struct {
	// Original event data (never modified)
	OriginalEvent *ical.Event    // The original iCal event
	Calendar      *ical.Calendar // Source calendar

	// Computed time information (may differ from original for recurring instances)
	ComputedStart timezone.LocalTime
	ComputedEnd   timezone.LocalTime

	// Instance metadata
	IsOriginal  bool   // True for regular events, false for recurring instances
	MasterUID   string // UID of master event
	InstanceUID string // Unique UID for this specific instance
}

// EventManagerInterface defines the interface for event operations to avoid circular imports
type EventManagerInterface interface {
	GetAllEventsForWeek(weekModel *WeekModel) ([]*EventInstance, error)
	GetTimedEventsForWeek(weekModel *WeekModel) ([]*EventInstance, error)
	GetAllDayEventsForWeek(weekModel *WeekModel) ([]*EventInstance, error)
	GetEventsForDayUnderCursor(weekModel *WeekModel) ([]*EventInstance, error)
	GetAllDayEventsForDayUnderCursor(weekModel *WeekModel) ([]*EventInstance, error)
	GetCalendarsForDisplay() []*ical.Calendar
	DeleteEvent(event *ical.Event) (error, tea.Cmd)
}

// RecurringEventOperationScope defines what scope of a recurring event operation (this/future/all)
type RecurringEventOperationScope int

const (
	OperationScopeThisOnly      RecurringEventOperationScope = iota // Edit/delete this occurrence only
	OperationScopeThisAndFuture                                     // Edit/delete this and all future occurrences
	OperationScopeAll                                               // Edit/delete all occurrences in the series
)

// RecurringEventOperationType defines the type of operation (edit vs delete)
type RecurringEventOperationType int

const (
	OperationTypeEdit   RecurringEventOperationType = iota // Edit operation
	OperationTypeDelete                                    // Delete operation
)

// RecurringEventOperation contains info about a pending recurring event operation
type RecurringEventOperation struct {
	OperationType RecurringEventOperationType  // Edit or delete
	TargetEvent   *ical.Event                  // The event being operated on
	InstanceDate  time.Time                    // The specific occurrence date
	IsRecurring   bool                         // Whether the event is actually recurring
	Scope         RecurringEventOperationScope // The selected scope (this/future/all) - set after dialog
}

// RecurringEventDialog represents the state of the "This/Future/All" dialog
type RecurringEventDialog struct {
	IsActive  bool
	Operation RecurringEventOperation
}

// AllDayEventGrid stores all-day events for a week with proper row assignment
type AllDayEventGrid struct {
	StartDate     time.Time
	EventRows     [AllDayCells][]AllDaySpanningEvent
	OverflowCount int  // Number of events that couldn't fit
	HasOverflow   bool // Whether there are more events than can be displayed
}

// AllDaySpanningEvent represents an all-day event with spanning information
type AllDaySpanningEvent struct {
	Event         *ical.Event    // For display/legacy compatibility
	EventInstance *EventInstance // For computed times and recurring events
	StartDay      int            // 0-6 (Sunday-Saturday)
	EndDay        int            // 0-6 (Sunday-Saturday)
	CalendarColor string         // calendar's color for consistency
}

type CellLayout []PositionedEvent

type DayLayout struct {
	Date        time.Time
	CellLayouts map[int]CellLayout
}

type WeekEventGrid struct {
	StartDate  time.Time
	DayLayouts [DaysPerWeek]DayLayout
}

type PositionedEvent struct {
	Instance      *EventInstance // EventInstance with computed times
	Column        int            // assigned column (0, 1, 2, etc.)
	IsStartCell   bool           // need to display title
	IsEndCell     bool           // handles edge cases of 2 adjacent events with same colors
	CalendarColor string         // calendar's X-APPLE-CALENDAR-COLOR for color hierarchy
}

// CollisionGroup represents events that overlap and need column assignment
type CollisionGroup []PositionedEvent
