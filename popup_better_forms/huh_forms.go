package popup_better_forms

import (
	"github.com/charmbracelet/huh/v2"
	"github.com/samuelstranges/chronos/types"
)

// SimpleAddEventForm creates a straightforward add event form with conditional groups
func SimpleAddEventForm(
	title, eventType, color, calendarID *string,
	startDate, startTime, duration, endDate *string,
	description, location, link *string,
	formComplexity, eventRecurrence *string,
	recurrenceType, recurrenceEnd, recurrenceCount, recurrenceUntil *string,
	calendars []types.CalendarInfo,
) *huh.Form {
	var groups []*huh.Group

	// Basic event details - always shown
	basicGroup := huh.NewGroup(
		CreateTitleSelector(title),
		CreateEventTypeSelector(eventType),
		CreateColorSelector(color),
		CreateCalendarIDSelector(calendars, calendarID),
	).Title("Event Details")
	groups = append(groups, basicGroup)

	// Timed event fields - only shown for timed events
	timedGroup := huh.NewGroup(
		CreateDateSelector("Start Date", startDate),
		CreateTimeSelector("Start Time", startTime),
		CreateDurationSelector(duration),
		CreateFormComplexitySelector(formComplexity),
		CreateEventRecurrenceSelector(eventRecurrence),
	).Title("Time Details").WithHideFunc(func() bool {
		return *eventType != "timed"
	})
	groups = append(groups, timedGroup)

	// All-day event fields - only shown for all-day events
	allDayGroup := huh.NewGroup(
		CreateDateSelector("Start Date", startDate),
		CreateEndDateSelector(endDate, startDate),
		CreateFormComplexitySelector(formComplexity),
		CreateEventRecurrenceSelector(eventRecurrence),
	).Title("Date Details").WithHideFunc(func() bool {
		return *eventType != "allday"
	})
	groups = append(groups, allDayGroup)

	// Advanced details - only shown if advanced selected
	advancedGroup := huh.NewGroup(
		CreateDescriptionSelector(description),
		CreateLocationSelector(location),
		CreateLinkSelector(link),
	).Title("Advanced Details").WithHideFunc(func() bool {
		return *formComplexity != "advanced"
	})
	groups = append(groups, advancedGroup)

	// Recurrence settings - only shown if recurring selected
	recurrenceGroup := huh.NewGroup(
		CreateRecurrenceTypeSelector(recurrenceType),
		CreateRecurrenceEndSelector(recurrenceEnd),
	).Title("Recurrence").WithHideFunc(func() bool {
		return *eventRecurrence != string(types.EventRecurrenceRecurring)
	})
	groups = append(groups, recurrenceGroup)

	// Recurrence count - only shown if end type is "count"
	countGroup := huh.NewGroup(
		CreateRecurrenceCountSelector(recurrenceCount, recurrenceEnd),
	).Title("Recurrence Count").WithHideFunc(func() bool {
		return *eventRecurrence != string(types.EventRecurrenceRecurring) || *recurrenceEnd != string(types.RecurrenceEndCount)
	})
	groups = append(groups, countGroup)

	// Recurrence until date - only shown if end type is "until"
	untilGroup := huh.NewGroup(
		CreateRecurrenceUntilSelector(recurrenceUntil, recurrenceEnd),
	).Title("End Date").WithHideFunc(func() bool {
		return *eventRecurrence != string(types.EventRecurrenceRecurring) || *recurrenceEnd != string(types.RecurrenceEndUntil)
	})
	groups = append(groups, untilGroup)

	return huh.NewForm(groups...) // Auto-sized
}

// SimpleEditEventForm creates a straightforward edit form for existing (timed) events
func SimpleEditEventForm(
	title, description, location, link, color *string,
	startDate, startTime, duration *string,
) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			CreateTitleSelector(title),
			CreateDateSelector("Start Date", startDate),
			CreateTimeSelector("Start Time", startTime),
			CreateDurationSelector(duration),
			CreateColorSelector(color),
		).Title("Basic Details"),

		huh.NewGroup(
			CreateDescriptionSelector(description),
			CreateLocationSelector(location),
			CreateLinkSelector(link),
		).Title("Additional Details"),
	) // Auto-sized
}

// SimpleEditAllDayEventForm creates a straightforward edit form for existing (all-day) events
func SimpleEditAllDayEventForm(
	title, description, location, link, color *string,
	startDate, endDate *string,
) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			CreateTitleSelector(title),
			CreateDateSelector("Start Date", startDate),
			CreateEndDateSelector(endDate, startDate),
			CreateColorSelector(color),
		).Title("Basic Details"),

		huh.NewGroup(
			CreateDescriptionSelector(description),
			CreateLocationSelector(location),
			CreateLinkSelector(link),
		).Title("Additional Details"),
	) // Auto-sized
}

// Simple single-field forms for quick edits

func SimpleTitleForm(title *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateTitleSelector(title)))
}

func SimpleColorForm(color *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateColorSelector(color)))
}

func SimpleDurationForm(duration *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateDurationSelector(duration)))
}

func SimpleDescriptionForm(description *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateDescriptionSelector(description)))
}

func SimpleLocationForm(location *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateLocationSelector(location)))
}

func SimpleLinkForm(link *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateLinkSelector(link)))
}

func SimpleStartTimeForm(startTime *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateTimeSelector("Start Time", startTime)))
}

func SimpleEndTimeForm(endTime *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(CreateTimeSelector("End Time", endTime)))
}
