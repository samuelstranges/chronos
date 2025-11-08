package navigation

import (
	"strings"
	
	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/types"
	"github.com/samuelstranges/chronos/util"
)

// NavigationCriteria defines what we're searching for
type NavigationCriteria interface {
	Matches(event types.PositionedEvent) bool
	Description() string
}

// EventStartCriteria matches events at their start position (w command)
type EventStartCriteria struct{}

func (c EventStartCriteria) Matches(event types.PositionedEvent) bool {
	return event.Instance != nil && event.IsStartCell
}

func (c EventStartCriteria) Description() string {
	return "event start"
}

// EventEndCriteria matches events at their end position (e command)
type EventEndCriteria struct{}

func (c EventEndCriteria) Matches(event types.PositionedEvent) bool {
	return event.Instance != nil && event.IsEndCell
}

func (c EventEndCriteria) Description() string {
	return "event end"
}

// SearchCriteria matches events based on search field/phrase (n/N commands)
type SearchCriteria struct {
	Field  string
	Phrase string
}

func (c SearchCriteria) Matches(event types.PositionedEvent) bool {
	if event.Instance == nil || event.Instance.OriginalEvent == nil {
		return false
	}
	
	// Only match at event start positions for search
	if !event.IsStartCell {
		return false
	}
	
	return c.matchesSearchCriteria(event.Instance.OriginalEvent, c.Field, c.Phrase)
}

func (c SearchCriteria) Description() string {
	if c.Field == "" || c.Field == "all" {
		return "search: " + c.Phrase
	}
	return "search " + c.Field + ": " + c.Phrase
}

// matchesSearchCriteria checks if an event matches the search field and phrase
func (c SearchCriteria) matchesSearchCriteria(event *ical.Event, field, phrase string) bool {
	phrase = strings.ToLower(phrase)
	field = strings.ToLower(field)
	
	switch field {
	case "title", "summary":
		return strings.Contains(strings.ToLower(util.GetEventTitle(event)), phrase)
	case "description":
		return strings.Contains(strings.ToLower(util.GetEventDescription(event)), phrase)
	case "location":
		return strings.Contains(strings.ToLower(util.GetEventLocation(event)), phrase)
	case "all", "":
		// Search all fields
		return strings.Contains(strings.ToLower(util.GetEventTitle(event)), phrase) ||
			strings.Contains(strings.ToLower(util.GetEventDescription(event)), phrase) ||
			strings.Contains(strings.ToLower(util.GetEventLocation(event)), phrase)
	default:
		return false
	}
}