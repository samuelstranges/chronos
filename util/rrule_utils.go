package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/timezone"
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
	var body string
	if idx := strings.Index(fullStr, "RRULE:"); idx != -1 {
		body = fullStr[idx+6:]
	} else {
		body = fullStr
	}
	// rrule-go hardcodes UNTIL as UTC ("…Z"), but Chronos writes DTSTART as
	// floating local. RFC 5545 §3.3.10 requires UNTIL's value type to match
	// DTSTART's, so rewrite a UTC UNTIL into floating local time.
	return RewriteRRuleUntilToFloating(body)
}

// RewriteRRuleUntilToFloating converts a UTC UNTIL token inside an RRULE value
// ("UNTIL=YYYYMMDDTHHMMSSZ") into floating local time ("UNTIL=YYYYMMDDTHHMMSS"),
// leaving every other token untouched. Values that aren't UTC datetimes
// (already floating, VALUE=DATE, etc.) pass through unchanged.
func RewriteRRuleUntilToFloating(rruleValue string) string {
	parts := strings.Split(rruleValue, ";")
	for i, part := range parts {
		name, value, ok := splitRRulePart(part)
		if !ok || name != "UNTIL" {
			continue
		}
		floating, ok := utcDateTimeToFloatingLocal(value)
		if !ok {
			continue
		}
		parts[i] = "UNTIL=" + floating
	}
	return strings.Join(parts, ";")
}

// splitRRulePart splits an RRULE token like "FREQ=WEEKLY" into name and value.
func splitRRulePart(token string) (name, value string, ok bool) {
	idx := strings.IndexByte(token, '=')
	if idx < 0 {
		return "", "", false
	}
	return token[:idx], token[idx+1:], true
}

// utcDateTimeToFloatingLocal converts "YYYYMMDDTHHMMSSZ" to the user's local
// wall-clock time formatted as "YYYYMMDDTHHMMSS". Returns false when the input
// isn't a UTC datetime (e.g. already floating, or VALUE=DATE).
func utcDateTimeToFloatingLocal(value string) (string, bool) {
	t, err := time.Parse("20060102T150405Z", value)
	if err != nil {
		return "", false
	}
	return timezone.NewLocalTime(t).Format("20060102T150405"), true
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

		// EXDATE may be: UTC (Z-suffixed), floating local time (no suffix, matches
		// a floating DTSTART), or date-only. Try each in turn.
		var date time.Time
		var err error
		if strings.HasSuffix(dateStr, "Z") {
			date, err = time.Parse("20060102T150405Z", dateStr)
		} else if len(dateStr) == 15 { // YYYYMMDDTHHMMSS
			date, err = time.ParseInLocation("20060102T150405", dateStr, timezone)
		} else {
			date, err = time.ParseInLocation("20060102", dateStr, timezone)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse EXDATE '%s': %w", dateStr, err)
		}

		// Normalize to the event's timezone so callers can compare uniformly
		dates = append(dates, date.In(timezone))
	}

	return dates, nil
}