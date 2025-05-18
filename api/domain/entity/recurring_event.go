package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/constant"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type RecurringEvent struct {
	CalendarID valueobject.CalendarID
	ID         valueobject.EventID
	Summary    string
	Recurrence string
	Start      *time.Time
	End        *time.Time
	Status     string
}

// NewCancelledRecurringEventFromEvent creates a new RecurringEvent from a Event.
// When an Event is updated to a RecurringEvent, the original Event is treated as cancelled.
// This function is used to represent that state.
func NewCancelledRecurringEventFromEvent(event Event) RecurringEvent {
	return RecurringEvent{
		CalendarID: event.CalendarID,
		ID:         event.ID,
		Summary:    event.Summary,
		Recurrence: "", // Recurrence is not set for cancelled events
		Start:      event.Start,
		End:        event.End,
		Status:     constant.EventStatusCancelled,
	}
}

func (e *RecurringEvent) Equals(other *RecurringEvent) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}

	return e.ID == other.ID &&
		e.CalendarID == other.CalendarID &&
		e.Summary == other.Summary &&
		e.Recurrence == other.Recurrence &&
		compareTime(e.Start, other.Start) &&
		compareTime(e.End, other.End) &&
		e.Status == other.Status
}
