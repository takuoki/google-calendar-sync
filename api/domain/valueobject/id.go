package valueobject

import "encoding/base64"

type CalendarID string

func (c CalendarID) ToChannelID() string {
	return base64.StdEncoding.EncodeToString([]byte(c))
}

type EventID string

func NewEventID(id string) *EventID {
	if id == "" {
		return nil
	}
	return pointer(EventID(id))
}

type ResourceID string

func pointer[T any](v T) *T {
	return &v
}
