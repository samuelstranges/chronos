package storage

import (
	"fmt"
	"time"

	"github.com/emersion/go-ical"
	"github.com/samuelstranges/chronos/util"
)

// ensureCalendarValid ensures calendar and all events have required properties
func ensureCalendarValid(calendar *ical.Calendar) error {
	// Ensure calendar has required properties
	if calendar.Props.Get(ical.PropProductID) == nil {
		calendar.Props.SetText(ical.PropProductID, "-//Chronos//Chronos Calendar//EN")
	}
	if calendar.Props.Get(ical.PropVersion) == nil {
		calendar.Props.SetText(ical.PropVersion, "2.0")
	}

	// Ensure all events have required properties
	var validationErr error
	util.ForEachEventInCalendar(calendar, func(event *ical.Component) bool {
		// DTSTAMP is required - set to current time if missing
		if event.Props.Get(ical.PropDateTimeStamp) == nil {
			event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
		}
		// UID is required for events
		if event.Props.Get(ical.PropUID) == nil {
			uid, uuidErr := util.GenerateUUID()
			if uuidErr != nil {
				validationErr = fmt.Errorf("failed to generate UUID: %w", uuidErr)
				return false // Stop iteration on error
			}
			event.Props.SetText(ical.PropUID, uid)
		}
		return true // Continue iteration
	})

	return validationErr
}

// ensureEventUID ensures a single event has a UID, generating one if missing
func ensureEventUID(event *ical.Event) error {
	if event.Props.Get(ical.PropUID) == nil {
		uid, err := util.GenerateUUID()
		if err != nil {
			return fmt.Errorf("failed to generate UUID: %w", err)
		}
		event.Props.SetText(ical.PropUID, uid)
	}
	return nil
}