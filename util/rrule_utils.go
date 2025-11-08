package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/teambition/rrule-go"
)

// ParseRRuleFromComponent extracts and parses RRULE from an iCal component
func ParseRRuleFromComponent(component *ical.Component) (*rrule.RRule, error) {
	rruleProp := component.Props.Get("RRULE")
	if rruleProp == nil {
		return nil, fmt.Errorf("component has no RRULE property")
	}

	return ParseRRuleString(rruleProp.Value)
}

// ParseRRuleString parses an RRULE string with proper unescaping
func ParseRRuleString(rruleStr string) (*rrule.RRule, error) {
	// Unescape iCal text escapes
	unescaped := UnescapeRRuleString(rruleStr)

	rule, err := rrule.StrToRRule(unescaped)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RRULE '%s': %w", unescaped, err)
	}

	return rule, nil
}

// UnescapeRRuleString handles iCal text unescaping
func UnescapeRRuleString(rruleStr string) string {
	result := rruleStr
	result = strings.ReplaceAll(result, "\\;", ";")
	result = strings.ReplaceAll(result, "\\,", ",")
	result = strings.ReplaceAll(result, "\\\\", "\\")
	return result
}

// ExtractRRuleStringFromRule gets just the RRULE part from a rule's string representation
func ExtractRRuleStringFromRule(rule *rrule.RRule) string {
	fullStr := rule.String()
	if idx := strings.Index(fullStr, "RRULE:"); idx != -1 {
		return fullStr[idx+6:] // Remove "RRULE:" prefix
	}
	return fullStr
}

// ParseExistingExdates parses a comma-separated EXDATE string into time slice
func ParseExistingExdates(exdateStr string, timezone *time.Location) ([]time.Time, error) {
	if exdateStr == "" {
		return nil, nil
	}

	// Unescape backslash-escaped commas that go-ical adds
	unescapedStr := strings.ReplaceAll(exdateStr, "\\,", ",")
	
	// Split comma-separated dates
	dateStrs := strings.Split(unescapedStr, ",")
	var dates []time.Time

	for _, dateStr := range dateStrs {
		dateStr = strings.TrimSpace(dateStr)
		if dateStr == "" {
			continue
		}

		// Parse iCal datetime format
		date, err := time.Parse("20060102T150405Z", dateStr)
		if err != nil {
			// Try parsing as date-only format
			date, err = time.Parse("20060102", dateStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse EXDATE '%s': %w", dateStr, err)
			}
		}

		// Convert to event's timezone
		dates = append(dates, date.In(timezone))
	}

	return dates, nil
}