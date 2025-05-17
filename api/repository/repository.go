package repository

import (
	"context"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type GoogleCalendarRepository interface {
	// events
	ListEventsWithAfter(ctx context.Context, calendarID valueobject.CalendarID, after time.Time) (
		events []entity.Event, recurringEvents []entity.RecurringEvent, nextSyncToken string, err error)

	ListEventsWithSyncToken(ctx context.Context, calendarID valueobject.CalendarID, syncToken string) (
		events []entity.Event, recurringEvents []entity.RecurringEvent, nextSyncToken string, err error)

	ListEventInstancesBetween(ctx context.Context, calendarID valueobject.CalendarID, eventID valueobject.EventID, from, to time.Time) (
		[]entity.Event, error)

	Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error)

	// channels
	StopWatch(ctx context.Context, channel entity.Channel) error
}

type DatabaseRepository interface {
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx DatabaseTransaction) error) error

	// calendars
	GetCalendar(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Calendar, error)
	ListCalendars(ctx context.Context) ([]entity.Calendar, error)
	GetRefreshToken(ctx context.Context, calendarID valueobject.CalendarID) (string, error)

	// recurring_events
	ListActiveRecurringEventsWithAfter(ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.RecurringEvent, error)

	// sync_histories
	GetLatestSyncToken(ctx context.Context, calendarID valueobject.CalendarID) (syncToken string, err error)
}

type DatabaseTransaction interface {
	// calendars
	LockCalendar(ctx context.Context, calendarID valueobject.CalendarID) error
	CreateCalendar(ctx context.Context, calendar entity.Calendar) error

	// recurring_events
	SyncRecurringEventAndInstancesWithAfter(ctx context.Context, recurringEvent entity.RecurringEvent, instances []entity.Event, after time.Time) (
		updatedCount int, err error)

	// events
	SyncEvents(ctx context.Context, calendarID valueobject.CalendarID, events []entity.Event) (updatedCount int, err error)

	// channel_histories
	ListActiveChannelHistoriesWithLock(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Channel, error)
	CreateChannelHistory(ctx context.Context, channel entity.Channel) error
	StopActiveChannels(ctx context.Context, calendarID valueobject.CalendarID) error

	// sync_histories
	CreateSyncHistory(
		ctx context.Context,
		calendarID valueobject.CalendarID,
		syncTime time.Time,
		nextSyncToken string,
		updatedEventCount int,
	) error
}
