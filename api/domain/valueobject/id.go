package valueobject

import (
	"fmt"
	"time"
)

type CalendarID string

func (c CalendarID) ToChannelID() string {
	return fmt.Sprintf("%s-%s", c, time.Now().Format("20060102150405"))
}

type ChannelID string

type EventID string
