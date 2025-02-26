package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type Event struct {
	ID         valueobject.EventID
	CalendarID valueobject.CalendarID
	Summary    string
	Start      *time.Time
	End        *time.Time
	Status     string
}

func (e *Event) Equals(other *Event) bool {
	if e == other {
		return true
	}
	if other == nil {
		return false
	}
	return e.ID == other.ID &&
		e.CalendarID == other.CalendarID &&
		e.Summary == other.Summary &&
		compareTime(e.Start, other.Start) &&
		compareTime(e.End, other.End) &&
		e.Status == other.Status
}

// compareTime is a helper function to compare two *time.Time values.
func compareTime(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 != nil && t2 != nil {
		return t1.Equal(*t2)
	}

	return false
}
