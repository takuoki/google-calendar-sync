package service

import "time"

const location = "Asia/Tokyo"

type Clock interface {
	Now() time.Time
	Today() time.Time
}

type clock struct{}

func NewClock() Clock {
	return &clock{}
}

func (c *clock) Now() time.Time {
	loc, _ := time.LoadLocation(location)
	return time.Now().In(loc)
}

func (c *clock) Today() time.Time {
	return c.Now().Truncate(24 * time.Hour)
}
