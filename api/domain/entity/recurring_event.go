package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type RecurringEvent struct {
	ID         valueobject.EventID
	CalendarID valueobject.CalendarID
	Summary    string
	Recurrence string
	Start      *time.Time
	End        *time.Time
	Status     string
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
