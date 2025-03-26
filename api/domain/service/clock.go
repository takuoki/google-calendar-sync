package service

import "time"

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Now() time.Time {
	return time.Now()
}

func (c *Clock) Today() time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return c.Now().In(loc).Truncate(24 * time.Hour)
}
