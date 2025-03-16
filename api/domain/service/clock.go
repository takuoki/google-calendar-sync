package service

import "time"

type Clock interface {
	Now() time.Time
	Today() time.Time
}

type clock struct{}

func NewClock() Clock {
	return &clock{}
}

func (c *clock) Now() time.Time {
	return time.Now()
}

func (c *clock) Today() time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return c.Now().In(loc).Truncate(24 * time.Hour)
}
