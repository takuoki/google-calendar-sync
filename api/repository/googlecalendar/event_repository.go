package googlecalendar

import (
	"context"
	"fmt"
	"time"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *googleCalendarRepository) ListEventsWithAfter(
	ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error) {
	return listEventsWithAfter(ctx, r.service, r.clockService, r.logger, calendarID, after)
}

func (r *googleCalendarWithOauthRepository) ListEventsWithAfter(
	ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get calendar service: %w", err)
	}

	return listEventsWithAfter(ctx, service, r.clockService, r.logger, calendarID, after)
}

func listEventsWithAfter(
	ctx context.Context, service *calendar.Service, clockService service.Clock, logger applog.Logger,
	calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error) {

	call := &eventsListCallWrapper{
		call: service.Events.List(string(calendarID)).Context(ctx).
			ShowDeleted(true).
			TimeMin(after.Format(time.RFC3339)),
	}

	return listEvents(ctx, clockService, logger, call, calendarID)
}

func (r *googleCalendarRepository) ListEventsWithSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {
	return listEventsWithSyncToken(ctx, r.service, r.clockService, r.logger, calendarID, syncToken)
}

func (r *googleCalendarWithOauthRepository) ListEventsWithSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get calendar service: %w", err)
	}

	return listEventsWithSyncToken(ctx, service, r.clockService, r.logger, calendarID, syncToken)
}

func listEventsWithSyncToken(ctx context.Context, service *calendar.Service, clockService service.Clock, logger applog.Logger,
	calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {

	call := &eventsListCallWrapper{
		call: service.Events.List(string(calendarID)).Context(ctx).
			SyncToken(syncToken),
	}

	return listEvents(ctx, clockService, logger, call, calendarID)
}

func (r *googleCalendarRepository) ListEventInstances(
	ctx context.Context, calendarID valueobject.CalendarID, eventID valueobject.EventID) ([]entity.Event, error) {
	return listEventInstances(ctx, r.service, r.clockService, r.logger, calendarID, eventID)
}

func (r *googleCalendarWithOauthRepository) ListEventInstances(
	ctx context.Context, calendarID valueobject.CalendarID, eventID valueobject.EventID) ([]entity.Event, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("fail to get calendar service: %w", err)
	}

	return listEventInstances(ctx, service, r.clockService, r.logger, calendarID, eventID)
}

func listEventInstances(ctx context.Context, service *calendar.Service, clockService service.Clock, logger applog.Logger,
	calendarID valueobject.CalendarID, eventID valueobject.EventID) ([]entity.Event, error) {

	call := &eventsInstancesCallWrapper{
		call: service.Events.Instances(string(calendarID), string(eventID)).Context(ctx).
			// 個別イベント化されていない子イベントを更新する場合は、全削除＆全登録のため、削除されたものは取得不要
			// 個別イベント化されたイベントは、通常の listEvents で差分取得されるため、ここでは考慮不要
			ShowDeleted(false),
	}

	// 子イベント取得時は差分取得ではないため、syncToken は不要
	events, _, err := listEvents(ctx, clockService, logger, call, calendarID)

	return events, err
}

func (r *googleCalendarRepository) Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {
	return watch(ctx, r.service, r.webhookBaseURL, r.clockService, calendarID)
}

func (r *googleCalendarWithOauthRepository) Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("fail to get calendar service: %w", err)
	}

	return watch(ctx, service, r.webhookBaseURL, r.clockService, calendarID)
}

func watch(ctx context.Context, service *calendar.Service, webhookBaseURL string, clockService service.Clock,
	calendarID valueobject.CalendarID) (*entity.Channel, error) {

	// TODO: ttl の設定を追加したい
	request := calendar.Channel{
		Id:      calendarID.ToChannelID(),
		Type:    "web_hook",
		Address: fmt.Sprintf("%s/%s/", webhookBaseURL, calendarID),
	}

	channel, err := service.Events.Watch(string(calendarID), &request).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("fail to watch: %w", err)
	}

	expiration, err := convertUnitTime(channel.Expiration)
	if err != nil {
		return nil, fmt.Errorf("fail to convert expiration time: %w", err)
	}

	return &entity.Channel{
		CalendarID: calendarID,
		ResourceID: valueobject.ResourceID(channel.ResourceId),
		StartTime:  clockService.Now(),
		Expiration: expiration,
	}, nil
}
