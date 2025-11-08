package popup_better_forms

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/huh/v2"
	"github.com/samuelstranges/chronos/types"
	util "github.com/samuelstranges/chronos/util"
)

func CreateColorSelector(colorValue *string) *huh.Input {
	return huh.NewInput().
		Title("Event Color").
		Placeholder("r, or FF0000").
		Value(colorValue).
		Inline(true).
		Validate(func(s string) error {
			if s != "" && util.ParseColor(s) == "" {
				return fmt.Errorf("invalid color")
			}
			return nil
		})
}

func CreateTitleSelector(titleValue *string) *huh.Input {
	return huh.NewInput().
		Title("Event Title").
		Placeholder("Enter event title...").
		Value(titleValue).
		Inline(true).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("title is required")
			}
			return nil
		})
}

func CreateDurationSelector(durationValue *string) *huh.Input {
	return huh.NewInput().
		Title("Duration").
		Placeholder("90").
		Value(durationValue).
		Inline(true).
		Validate(func(s string) error {
			intDuration, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be valid number")
			}
			if intDuration <= 0 || intDuration > types.MinsPerDay {
				return fmt.Errorf("must be valid number 1-%d", types.MinsPerDay)
			}
			return nil
		})
}

func CreateCalendarIDSelector(calendarInfos []types.CalendarInfo, selectedCalendarID *string) huh.Field {
	// Default to first calendar if available
	if len(calendarInfos) > 0 && *selectedCalendarID == "" {
		*selectedCalendarID = calendarInfos[0].ID
	}

	// Create options from calendar infos
	options := make([]huh.Option[string], len(calendarInfos))
	for i, info := range calendarInfos {
		options[i] = huh.NewOption(info.Name, info.ID)
	}

	// If no calendars, create a default option
	if len(options) == 0 {
		options = []huh.Option[string]{
			huh.NewOption("No Calendars Available", ""),
		}
	}

	return huh.NewSelect[string]().
		Title("Calendar").
		Options(options...).
		Value(selectedCalendarID)
}

func CreateDateSelector(title string, dateValue *string) *huh.Input {
	return huh.NewInput().
		Title(title).
		Placeholder("2024-01-15").
		Value(dateValue).
		Inline(true).
		Validate(func(s string) error {
			_, err := time.Parse("2006-01-02", s)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD")
			}
			return nil
		})
}

func CreateTimeSelector(title string, timeValue *string) *huh.Input {
	return huh.NewInput().
		Title(title).
		Placeholder("14:30").
		Value(timeValue).
		Inline(true).
		Validate(func(s string) error {
			_, err := time.Parse("15:04", s)
			if err != nil {
				return fmt.Errorf("invalid time format, use HH:MM")
			}
			return nil
		})
}

// CreateEndDateSelector creates an end date selector with validation against start date
func CreateEndDateSelector(endDateValue *string, startDateValue *string) *huh.Input {
	return huh.NewInput().
		Title("End Date").
		Placeholder("2024-01-15").
		Value(endDateValue).
		Inline(true).
		Validate(func(s string) error {
			// Parse end date
			endDate, err := time.Parse("2006-01-02", s)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD")
			}

			// Parse start date
			startDate, err := time.Parse("2006-01-02", *startDateValue)
			if err != nil {
				// If start date is invalid, just validate end date format
				return nil
			}

			// Ensure end date >= start date
			if endDate.Before(startDate) {
				return fmt.Errorf("end date must be >= start date")
			}

			return nil
		})
}

func CreateDescriptionSelector(descriptionValue *string) huh.Field {
	return huh.NewText().
		Title("Description").
		Placeholder("Event description...").
		Value(descriptionValue).
		WithHeight(3)
}

func CreateLocationSelector(locationValue *string) *huh.Input {
	return huh.NewInput().
		Title("Location").
		Placeholder("Conference Room A, 123 Main St...").
		Value(locationValue).
		Inline(true)
}

func CreateLinkSelector(linkValue *string) *huh.Input {
	return huh.NewInput().
		Title("Link").
		Placeholder("https://example.com or tel:+1234567890").
		Value(linkValue).
		Inline(true).
		Validate(func(s string) error {
			if s == "" {
				return nil // Optional field
			}
			// Allow any URL scheme - no restrictions
			return nil
		})
}

// Recurring event selectors

func CreateRecurrenceTypeSelector(recurrenceTypeValue *string) *huh.Select[string] {
	options := []huh.Option[string]{
		huh.NewOption("Daily", "daily"),
		huh.NewOption("Weekly", "weekly"),
		huh.NewOption("Monthly", "monthly"),
		huh.NewOption("Yearly", "yearly"),
	}

	return huh.NewSelect[string]().
		Title("Repeat").
		Options(options...).
		Value(recurrenceTypeValue)
}

func CreateRecurrenceEndSelector(recurrenceEndValue *string) *huh.Select[string] {
	options := []huh.Option[string]{
		huh.NewOption("Never", "never"),
		huh.NewOption("After count", "count"),
		huh.NewOption("Until date", "until"),
	}

	return huh.NewSelect[string]().
		Title("End").
		Options(options...).
		Value(recurrenceEndValue)
}

func CreateRecurrenceCountSelector(recurrenceCountValue *string, recurrenceEnd *string) *huh.Input {
	return huh.NewInput().
		TitleFunc(func() string {
			if *recurrenceEnd == string(types.RecurrenceEndCount) {
				return "Number of occurrences"
			}
			return "(Count not needed)"
		}, recurrenceEnd).
		Placeholder("10").
		Value(recurrenceCountValue).
		Inline(true).
		Validate(func(s string) error {
			if *recurrenceEnd != string(types.RecurrenceEndCount) {
				return nil // Skip validation if not using count
			}
			if s == "" {
				return fmt.Errorf("count is required when ending after count")
			}
			if count, err := strconv.Atoi(s); err != nil || count < 1 || count > 1000 {
				return fmt.Errorf("count must be between 1 and 1000")
			}
			return nil
		})
}

func CreateRecurrenceUntilSelector(recurrenceUntilValue *string, recurrenceEnd *string) *huh.Input {
	return huh.NewInput().
		TitleFunc(func() string {
			if *recurrenceEnd == string(types.RecurrenceEndUntil) {
				return "End date"
			}
			return "(End date not needed)"
		}, recurrenceEnd).
		Placeholder("2024-12-31").
		Value(recurrenceUntilValue).
		Inline(true).
		Validate(func(s string) error {
			if *recurrenceEnd != string(types.RecurrenceEndUntil) {
				return nil // Skip validation if not using until
			}
			if s == "" {
				return fmt.Errorf("end date is required when ending on date")
			}
			if _, err := time.Parse("2006-01-02", s); err != nil {
				return fmt.Errorf("date must be in format YYYY-MM-DD")
			}
			return nil
		})
}

// Dynamic form control selectors

func CreateFormComplexitySelector(formComplexityValue *string) *huh.Select[string] {
	options := []huh.Option[string]{
		huh.NewOption("Simple (basic fields)", "simple"),
		huh.NewOption("Advanced (all fields)", "advanced"),
	}

	return huh.NewSelect[string]().
		Title("Form complexity").
		Options(options...).
		Value(formComplexityValue)
}

func CreateEventTypeSelector(eventTypeValue *string) *huh.Select[string] {
	options := []huh.Option[string]{
		huh.NewOption("Timed event", "timed"),
		huh.NewOption("All-day event", "allday"),
	}

	return huh.NewSelect[string]().
		Title("Event type").
		Options(options...).
		Value(eventTypeValue)
}

func CreateEventRecurrenceSelector(eventRecurrenceValue *string) *huh.Select[string] {
	options := []huh.Option[string]{
		huh.NewOption("One-time event", "onetime"),
		huh.NewOption("Recurring event", "recurring"),
	}

	return huh.NewSelect[string]().
		Title("Recurrence").
		Options(options...).
		Value(eventRecurrenceValue)
}

// Sub-dynamic recurrence selectors

func CreateRecurrenceWeekSelector(recurrenceWeekValue *string, recurrencePattern *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		TitleFunc(func() string {
			if *recurrencePattern == string(types.RecurrencePatternWeekday) {
				return "Which week"
			}
			return "Week position"
		}, recurrencePattern).
		OptionsFunc(func() []huh.Option[string] {
			if *recurrencePattern == string(types.RecurrencePatternWeekday) {
				return []huh.Option[string]{
					huh.NewOption("First week", "first"),
					huh.NewOption("Second week", "second"),
					huh.NewOption("Third week", "third"),
					huh.NewOption("Fourth week", "fourth"),
					huh.NewOption("Last week", "last"),
				}
			}
			return []huh.Option[string]{
				huh.NewOption("(Not applicable)", "none"),
			}
		}, recurrencePattern).
		Value(recurrenceWeekValue)
}
