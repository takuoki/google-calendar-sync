package service

import "time"

const location = "Asia/Tokyo"

type Clock interface {
	Now() time.Time
	Today() time.Time
}

type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	loc, _ := time.LoadLocation(location)
	return time.Now().In(loc)
}

func (c *SystemClock) Today() time.Time {
	return c.Now().Truncate(24 * time.Hour)
}
