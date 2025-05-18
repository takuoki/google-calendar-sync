package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/constant"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type Event struct {
	CalendarID       valueobject.CalendarID
	ID               valueobject.EventID
	RecurringEventID *valueobject.EventID
	Summary          string
	Start            *time.Time
	End              *time.Time
	Status           string
}

// NewEventFromRecurringEvent creates a new Event from a RecurringEvent.
// When a RecurringEvent is updated to a normal Event, the original RecurringEvent is treated as cancelled.
// This function is used to represent that state.
func NewCancelledEventFromRecurringEvent(recurringEvent RecurringEvent) Event {
	return Event{
		CalendarID:       recurringEvent.CalendarID,
		ID:               recurringEvent.ID,
		RecurringEventID: nil,
		Summary:          recurringEvent.Summary,
		Start:            recurringEvent.Start,
		End:              recurringEvent.End,
		Status:           constant.EventStatusCancelled,
	}
}

func (e *Event) Equals(other *Event) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}

	return e.CalendarID == other.CalendarID &&
		e.ID == other.ID &&
		comparePointer(e.RecurringEventID, other.RecurringEventID) &&
		e.Summary == other.Summary &&
		compareTime(e.Start, other.Start) &&
		compareTime(e.End, other.End) &&
		e.Status == other.Status
}
