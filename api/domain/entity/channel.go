package entity

import (
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type Channel struct {
	ID         valueobject.ChannelID
	CalendarID valueobject.CalendarID
	Expiration time.Time
}
