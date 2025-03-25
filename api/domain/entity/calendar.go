package entity

import "github.com/takuoki/google-calendar-sync/api/domain/valueobject"

type Calendar struct {
	ID           valueobject.CalendarID
	Name         string
	RefreshToken *string
}
