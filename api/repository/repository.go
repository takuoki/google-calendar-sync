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
		events []entity.Event, nextSyncToken string, err error)

	ListEventsWithSyncToken(ctx context.Context, calendarID valueobject.CalendarID, syncToken string) (
		events []entity.Event, nextSyncToken string, err error)

	Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error)

	// channels
	StopWatch(ctx context.Context, channel entity.Channel) error
}

type DatabaseRepository interface {
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx DatabaseTransaction) error) error

	// calendars
	GetCalendar(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Calendar, error)

	// sync_histories
	GetLatestSyncToken(ctx context.Context, calendarID valueobject.CalendarID) (syncToken string, err error)
}

type DatabaseTransaction interface {
	// calendars
	CreateCalendar(ctx context.Context, calendar entity.Calendar) error

	// events
	SyncEvents(ctx context.Context, events []entity.Event) (updatedCount int, err error)

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
