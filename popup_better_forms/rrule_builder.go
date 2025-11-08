package popup_better_forms

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/types"
	"github.com/teambition/rrule-go"
)

// BuildRRuleFromFormData creates an RRULE string from form pointer values
// This matches the parameter style used in SimpleAddEventForm
func BuildRRuleFromFormData(
	recurrenceType, recurrenceEnd, recurrenceCount, recurrenceUntil *string,
	startDate *string,
	location *time.Location,
) (string, error) {
	// Check if we have recurrence data
	if recurrenceType == nil || *recurrenceType == "" {
		return "", nil
	}

	// Parse the start date for DTSTART
	if startDate == nil || *startDate == "" {
		return "", fmt.Errorf("start date is required for recurring events")
	}

	startTime, err := time.ParseInLocation("2006-01-02", *startDate, location)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	// Build base options
	options := buildBaseOptions(startTime, *recurrenceType)
	if options == nil {
		return "", fmt.Errorf("unsupported recurrence type: %s", *recurrenceType)
	}

	// Add end condition options
	err = addEndOptions(options, recurrenceEnd, recurrenceCount, recurrenceUntil, location)
	if err != nil {
		return "", err
	}

	// Generate the rule
	return generateRuleString(*options)
}

// BuildRRuleWithAdvancedPattern creates RRULE with advanced patterns for monthly/yearly events
// This is for future advanced forms that might include weekday patterns
func BuildRRuleWithAdvancedPattern(
	recurrenceType, recurrenceEnd, recurrenceCount, recurrenceUntil *string,
	recurrencePattern, recurrenceWeek *string, // Advanced pattern options
	startDate *string,
	location *time.Location,
) (string, error) {
	// Start with basic RRULE
	basicRRule, err := BuildRRuleFromFormData(recurrenceType, recurrenceEnd, recurrenceCount, recurrenceUntil, startDate, location)
	if err != nil || basicRRule == "" {
		return basicRRule, err
	}

	// If no advanced pattern, return basic rule
	if recurrencePattern == nil || *recurrencePattern == "" {
		return basicRRule, nil
	}

	// Parse start date again for pattern calculations
	startTime, err := time.ParseInLocation("2006-01-02", *startDate, location)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	// Build options with pattern
	options := buildBaseOptions(startTime, *recurrenceType)
	if options == nil {
		return "", fmt.Errorf("unsupported recurrence type: %s", *recurrenceType)
	}

	// Add pattern-specific options
	addPatternOptions(options, recurrenceType, recurrencePattern, recurrenceWeek, startTime)

	// Add end conditions
	err = addEndOptions(options, recurrenceEnd, recurrenceCount, recurrenceUntil, location)
	if err != nil {
		return "", err
	}

	return generateRuleString(*options)
}

// buildBaseOptions creates base ROption with frequency
func buildBaseOptions(startTime time.Time, recurrenceType string) *rrule.ROption {
	options := &rrule.ROption{
		Dtstart: startTime,
	}

	switch recurrenceType {
	case "daily":
		options.Freq = rrule.DAILY
	case "weekly":
		options.Freq = rrule.WEEKLY
	case "monthly":
		options.Freq = rrule.MONTHLY
	case "yearly":
		options.Freq = rrule.YEARLY
	default:
		return nil
	}

	return options
}

// addPatternOptions adds weekday pattern options for monthly/yearly recurrence
func addPatternOptions(options *rrule.ROption, recurrenceType, recurrencePattern, recurrenceWeek *string, startTime time.Time) {
	// Only apply patterns to monthly/yearly with weekday pattern
	if (recurrenceType == nil || (*recurrenceType != "monthly" && *recurrenceType != "yearly")) ||
		(recurrencePattern == nil || *recurrencePattern != "weekday") {
		return
	}

	// Convert Go weekday to rrule weekday
	rruleWeekday := convertWeekday(startTime.Weekday())

	// Determine which week occurrence to use
	var weekSelection string
	if recurrenceWeek != nil {
		weekSelection = *recurrenceWeek
	}
	weekOccurrence := calculateWeekOccurrence(startTime, weekSelection)

	// Set the BYDAY pattern
	options.Byweekday = []rrule.Weekday{rruleWeekday.Nth(weekOccurrence)}
}

// convertWeekday maps Go time.Weekday to rrule.Weekday
func convertWeekday(weekday time.Weekday) rrule.Weekday {
	switch weekday {
	case time.Sunday:
		return rrule.SU
	case time.Monday:
		return rrule.MO
	case time.Tuesday:
		return rrule.TU
	case time.Wednesday:
		return rrule.WE
	case time.Thursday:
		return rrule.TH
	case time.Friday:
		return rrule.FR
	case time.Saturday:
		return rrule.SA
	default:
		return rrule.MO // fallback
	}
}

// calculateWeekOccurrence determines which week of the month (1st, 2nd, 3rd, 4th, last)
func calculateWeekOccurrence(startTime time.Time, userSelection string) int {
	// If user explicitly selected, use that
	if userSelection != "" {
		switch userSelection {
		case "first":
			return 1
		case "second":
			return 2
		case "third":
			return 3
		case "fourth":
			return 4
		case "last":
			return -1
		}
	}

	// Auto-detect based on start date
	weekOfMonth := (startTime.Day()-1)/types.DaysPerWeek + 1

	// Check if this is the last occurrence of this weekday in the month
	nextWeek := startTime.AddDate(0, 0, types.DaysPerWeek)
	if nextWeek.Month() != startTime.Month() {
		return -1 // Last occurrence
	}

	return weekOfMonth
}

// addEndOptions adds end condition options (count, until, never)
func addEndOptions(options *rrule.ROption, recurrenceEnd, recurrenceCount, recurrenceUntil *string, location *time.Location) error {
	if recurrenceEnd == nil {
		return nil
	}

	switch *recurrenceEnd {
	case "count":
		if recurrenceCount != nil && *recurrenceCount != "" {
			count, err := strconv.Atoi(*recurrenceCount)
			if err != nil {
				return fmt.Errorf("invalid recurrence count: %w", err)
			}
			options.Count = count
		}
	case "until":
		if recurrenceUntil != nil && *recurrenceUntil != "" {
			untilDate, err := time.ParseInLocation("2006-01-02", *recurrenceUntil, location)
			if err != nil {
				return fmt.Errorf("invalid until date: %w", err)
			}
			// Set to end of day for until date
			options.Until = time.Date(untilDate.Year(), untilDate.Month(), untilDate.Day(), 23, 59, 59, 0, location)
		}
	case "never":
		// No additional options needed - rule will repeat indefinitely
	}

	return nil
}

// generateRuleString creates the final RRULE string from options
func generateRuleString(options rrule.ROption) (string, error) {
	// Create the rule
	rule, err := rrule.NewRRule(options)
	if err != nil {
		return "", fmt.Errorf("failed to create rrule: %w", err)
	}

	// Get the full string and extract just the RRULE line
	fullStr := rule.String()

	// Find and extract just the RRULE line
	// The rrule library returns iCal format like:
	// DTSTART;TZID=Local:20250812T000000
	// RRULE:FREQ=DAILY
	lines := strings.Split(fullStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that are exactly RRULE properties (not SUMMARY:RRULE: etc)
		if line == "RRULE:" || (strings.HasPrefix(line, "RRULE:") && len(line) > 6) {
			extracted := line[6:] // Remove "RRULE:" prefix
			return extracted, nil
		}
	}

	// Fallback: no RRULE line found
	return "", fmt.Errorf("no RRULE found in generated string")
}