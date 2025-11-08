// Package search provides event search functionality across calendars.
package search_parser

import (
	"strings"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
)

const (
	searchFieldParts = 2
)


// ParseSearchInput parses a search input string in the format "/FIELD/phrase/"
// Returns the field, phrase, and whether the parse was successful
func ParseSearchInput(input string) (string, string, bool) {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Must start with /
	if !strings.HasPrefix(input, "/") {
		return "", "", false
	}

	// Remove the leading /
	input = input[1:]

	// If empty after removing /, not valid yet
	if input == "" {
		return "", "", false
	}

	// Find the next /
	parts := strings.SplitN(input, "/", searchFieldParts)
	if len(parts) != searchFieldParts {
		// Still building the command, return partial state
		// This handles cases like "/SUMM" or "/SUMMARY"
		field := strings.ToUpper(parts[0])
		return field, "", true
	}

	field := strings.ToUpper(parts[0])
	phrase := parts[1]

	return field, phrase, true
}



// IsEventInSearchResultsByInstance checks if a given event is in the EventInstance-based search results
func IsEventInSearchResultsByInstance(event *ical.Event, instances []*types.EventInstance) bool {
	if event == nil {
		return false
	}

	eventUID := ""
	if uidProp := event.Props.Get(ical.PropUID); uidProp != nil {
		eventUID = uidProp.Value
	}

	for _, instance := range instances {
		if instance == nil || instance.OriginalEvent == nil {
			continue
		}

		resultUID := ""
		if uidProp := instance.OriginalEvent.Props.Get(ical.PropUID); uidProp != nil {
			resultUID = uidProp.Value
		}

		// Compare by UID if available, otherwise compare by pointer
		if eventUID != "" && resultUID != "" {
			if eventUID == resultUID {
				return true
			}
		} else if instance.OriginalEvent == event {
			return true
		}
	}

	return false
}
