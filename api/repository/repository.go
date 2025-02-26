package repository

import (
	"context"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

type GoogleCalendarRepository interface {
	ListEvents(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Event, string, error)
	ListEventsWithSyncToken(ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error)
	Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error)
	StopWatch(ctx context.Context, channelID valueobject.ChannelID) error
}

type DatabaseRepository interface {
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx DatabaseTransaction) error) error

	// calendars
	GetCalendar(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Calendar, error)

	// sync_histories
	GetLatestSyncToken(ctx context.Context, calendarID valueobject.CalendarID) (string, error)
}

type DatabaseTransaction interface {
	// calendars
	CreateCalendar(ctx context.Context, calendar entity.Calendar) error

	// channels
	ListActiveChannels(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Channel, error)
	CreateChannel(ctx context.Context, channel entity.Channel) error
	StopChannel(ctx context.Context, channelID valueobject.ChannelID) error

	// events
	SyncEvents(ctx context.Context, events []entity.Event) (int, error)

	// sync_histories
	CreateSyncHistory(
		ctx context.Context,
		calendarID valueobject.CalendarID,
		syncTime time.Time,
		nextSyncToken string,
		updatedEventCount int,
	) error
}
