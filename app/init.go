package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/huh/v2"
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/config"
	"github.com/samuelstranges/chronos/ical_crud"
	keybinds "github.com/samuelstranges/chronos/keybinds"
	keyhints "github.com/samuelstranges/chronos/keybinds_hints"
	"github.com/samuelstranges/chronos/storage"
	types "github.com/samuelstranges/chronos/types"
	util "github.com/samuelstranges/chronos/util"
	"github.com/samuelstranges/chronos/week_view_shared"
)

// Model represents the main application state
type Model struct {
	WeekModel            types.WeekModel
	EventManager         *ical_crud.EventManager
	Form                 *huh.Form
	ShowForm             bool
	RecurringOperation   *types.RecurringEventOperation // Store current recurring operation
	Quitting             bool
	VimState             *keybinds.VimState
	VimRegistry          *keybinds.CommandRegistry
	KeyHints             *keyhints.KeyHintSystem
	WaitingForKeyTimeout bool
	SyncInterval         int // CalDAV background sync interval in seconds (0 = disabled)
}

// Init implements the tea.Model interface
func (m Model) Init() tea.Cmd {
	// Start background sync ticker if enabled
	if m.SyncInterval > 0 {
		return tea.Tick(time.Duration(m.SyncInterval)*time.Second, func(t time.Time) tea.Msg {
			return types.BackgroundSyncMsg{}
		})
	}
	return nil
}

// NewModel creates and initializes a new application model
func NewModel(bgHex string) Model {
	weekStart := util.GetWeekStartingOn(time.Now(), time.Sunday)

	// Initialize cursor at current time
	now := time.Now()
	currentDay := util.GetDayOfWeek(now, weekStart)
	currentCell := util.TimeToCell(now, types.Zoom30Min)

	weekModel := types.WeekModel{
		CurrentlyViewedWeek:           weekStart,
		Cursor:                        types.CursorPositionInWeek{Day: int(currentDay), Cell: currentCell, EventColumn: 0},
		CurrentZoom:                   types.Zoom30Min,
		CachedBackgroundColor:         bgHex,
		CachedBackgroundSelectedColor: util.BrightenColor(bgHex, types.BrightenBackgroundPercent),
		ShowAllDayGrid:                true, // Show all-day grid by default
		EventTextBlack:                true, // Default to black text
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize storage based on config
	var calStorage storage.CalendarStorage
	switch cfg.StorageType {
	case config.StorageTypeCalDAV:
		// Resolve password from configured source
		password, err := cfg.CalDAV.GetPassword()
		if err != nil {
			// Failed to get password - PANIC for now to see the error
			panic(fmt.Sprintf("CalDAV password resolution failed: %v", err))
		}

		// Convert config format to storage format
		caldavCfg := &storage.CalDAVConfig{
			Enabled:         true,
			ServerURL:       cfg.CalDAV.ServerURL,
			Username:        cfg.CalDAV.Username,
			Password:        password, // Use resolved password
			CalendarHomeURL: cfg.CalDAV.CalendarHomeURL,
			SyncInterval:    cfg.CalDAV.SyncInterval, // Use configured sync interval
		}
		calStorage, err = storage.NewCalDAVStorage(caldavCfg)
		if err != nil {
			// Failed to connect to CalDAV - PANIC for now to see the error
			panic(fmt.Sprintf("CalDAV connection failed: %v", err))
		}
	default:
		// Default to file storage
		calStorage = storage.NewFileStorage()
	}

	// Try to load calendars
	calendarMap, loadErr := calStorage.LoadCalendars()
	var eventMgr *ical_crud.EventManager

	if loadErr == nil && len(calendarMap) > 0 {
		// Successfully loaded calendars - use them directly with stable IDs!
		eventMgr = ical_crud.New(calendarMap, calStorage)
	} else {
		// No calendars found - start with empty calendar map
		emptyCalendarMap := make(map[string]*ical.Calendar)
		eventMgr = ical_crud.New(emptyCalendarMap, calStorage)
	}

	// Initialize cache with default values (will be updated on first WindowSizeMsg)
	weekModel.CachedCellWidth, weekModel.CachedContentWidth = week_view_shared.CalculateCellWidth(80)

	// Refresh display with initial calendar data
	refreshErr := refreshEventGridsWithEventManager(&weekModel, eventMgr)
	util.ShowIfError(&weekModel, 5, refreshErr, "Failed to load initial events")

	// Initialize vim system
	vimState := keybinds.NewVimState()
	vimRegistry := keybinds.NewCommandRegistry()

	// Initialize key hints system with vim registry
	vimAdapter := keybinds.NewKeyhintAdapter(vimRegistry)
	keyHints := keyhints.NewKeyHintSystem(vimAdapter)

	// Determine sync interval (CalDAV only)
	syncInterval := 0
	if cfg.StorageType == config.StorageTypeCalDAV && cfg.CalDAV != nil {
		syncInterval = cfg.CalDAV.SyncInterval
	}

	model := Model{
		WeekModel:    weekModel,
		EventManager: eventMgr,
		VimState:     vimState,
		VimRegistry:  vimRegistry,
		KeyHints:     keyHints,
		SyncInterval: syncInterval,
	}

	// Set up refresh callback now that we have the model
	// setupRefreshCallback() - REMOVED: Now using tea.Cmd pattern

	return model
}
