package valueobject

import "encoding/base64"

type CalendarID string

func (c CalendarID) ToChannelID() string {
	return base64.StdEncoding.EncodeToString([]byte(c))
}

type EventID string

type ResourceID string
