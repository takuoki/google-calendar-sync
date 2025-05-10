package service

import (
	"fmt"
	"time"
)

type Clock interface {
	Now() time.Time
	Today() time.Time
	ConvertDate(date string) (time.Time, error)
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
	return time.Now().Truncate(time.Millisecond).In(c.location)
}

func (c *SystemClock) Today() time.Time {
	return c.Now().Truncate(24 * time.Hour)
}

func (c *SystemClock) ConvertDate(date string) (time.Time, error) {
	return convertDate(date, c.location)
}

type MockClock struct {
	fixedTime time.Time
}

func NewMockClock() *MockClock {
	return &MockClock{
		fixedTime: time.Now().Truncate(time.Millisecond),
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

func (c *MockClock) ConvertDate(date string) (time.Time, error) {
	return convertDate(date, time.UTC)
}

func convertDate(date string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("fail to parse date: %w", err)
	}
	return t, nil
}
