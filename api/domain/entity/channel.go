package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type Channel struct {
	CalendarID valueobject.CalendarID
	ResourceID valueobject.ResourceID
	StartTime  time.Time
	Expiration time.Time
	IsStopped  bool
}
