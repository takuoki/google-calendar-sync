package service

import (
	"fmt"
	"time"
)

type Clock interface {
	Now() time.Time
	Today() time.Time
}

type SystemClock struct {
	location *time.Location
}

func NewSystemClock(location string) (*SystemClock, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, fmt.Errorf("fail to load location: %w", err)
	}

	return &SystemClock{
		location: loc,
	}, nil
}

func (c *SystemClock) Now() time.Time {
	return time.Now().In(c.location)
}

func (c *SystemClock) Today() time.Time {
	return c.Now().Truncate(24 * time.Hour)
}

type MockClock struct {
	fixedTime time.Time
}

func NewMockClock() *MockClock {
	return &MockClock{
		fixedTime: time.Now(),
	}
}

func (c *MockClock) SetFixedTime(t time.Time) {
	c.fixedTime = t
}

func (c *MockClock) Now() time.Time {
	return c.fixedTime
}

func (c *MockClock) Today() time.Time {
	return c.fixedTime.Truncate(24 * time.Hour)
}
