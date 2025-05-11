package usecase_test

import (
	"context"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type GoogleCalendarRepositoryMock struct {
	ListEventsWithAfterFunc     func(ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error)
	ListEventsWithSyncTokenFunc func(ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error)
	ListEventInstancesFunc      func(ctx context.Context, calendarID valueobject.CalendarID, eventID valueobject.EventID) ([]entity.Event, error)
	WatchFunc                   func(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error)
	StopWatchFunc               func(ctx context.Context, channel entity.Channel) error
}

func (m *GoogleCalendarRepositoryMock) ListEventsWithAfter(ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error) {
	if m.ListEventsWithAfterFunc != nil {
		return m.ListEventsWithAfterFunc(ctx, calendarID, after)
	}
	return nil, "", nil
}

func (m *GoogleCalendarRepositoryMock) ListEventsWithSyncToken(ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {
	if m.ListEventsWithSyncTokenFunc != nil {
		return m.ListEventsWithSyncTokenFunc(ctx, calendarID, syncToken)
	}
	return nil, "", nil
}

func (m *GoogleCalendarRepositoryMock) ListEventInstances(ctx context.Context, calendarID valueobject.CalendarID, eventID valueobject.EventID) ([]entity.Event, error) {
	if m.ListEventInstancesFunc != nil {
		return m.ListEventInstancesFunc(ctx, calendarID, eventID)
	}
	return nil, nil
}

func (m *GoogleCalendarRepositoryMock) Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {
	if m.WatchFunc != nil {
		return m.WatchFunc(ctx, calendarID)
	}
	return nil, nil
}

func (m *GoogleCalendarRepositoryMock) StopWatch(ctx context.Context, channel entity.Channel) error {
	if m.StopWatchFunc != nil {
		return m.StopWatchFunc(ctx, channel)
	}
	return nil
}
