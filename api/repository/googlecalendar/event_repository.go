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
	call := r.service.Events.List(string(calendarID)).Context(ctx).TimeMin(after.Format(time.RFC3339))
	return listEvents(ctx, call, calendarID, r.logger)
}

func (r *googleCalendarWithOauthRepository) ListEventsWithAfter(
	ctx context.Context, calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, string, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get calendar service: %w", err)
	}

	call := service.Events.List(string(calendarID)).Context(ctx).TimeMin(after.Format(time.RFC3339))
	return listEvents(ctx, call, calendarID, r.logger)
}

func (r *googleCalendarRepository) ListEventsWithSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {
	call := r.service.Events.List(string(calendarID)).Context(ctx).SyncToken(syncToken)
	return listEvents(ctx, call, calendarID, r.logger)
}

func (r *googleCalendarWithOauthRepository) ListEventsWithSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {

	service, err := r.getCalendarService(ctx, calendarID)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get calendar service: %w", err)
	}

	call := service.Events.List(string(calendarID)).Context(ctx).SyncToken(syncToken)
	return listEvents(ctx, call, calendarID, r.logger)
}

func listEvents(ctx context.Context, baseCall *calendar.EventsListCall,
	calendarID valueobject.CalendarID, logger applog.Logger) ([]entity.Event, string, error) {

	pageToken := ""
	syncToken := ""
	res := []entity.Event{}
	for syncToken == "" { // 最後のページまで取得すると必ず値が入る

		// TODO: ctx のキャンセルを確認する

		var call *calendar.EventsListCall
		if pageToken == "" {
			call = baseCall
		} else {
			call = baseCall.PageToken(pageToken)
		}

		events, err := call.Do()
		if err != nil {
			return nil, "", fmt.Errorf("fail to list events: %w", err)
		}

		for _, item := range events.Items {

			logger.Debugf(ctx, "list events: item=%+v", item)

			start, err := convertDateTime(item.Start)
			if err != nil {
				return nil, "", fmt.Errorf("fail to convert start datetime: %w", err)
			}
			end, err := convertDateTime(item.End)
			if err != nil {
				return nil, "", fmt.Errorf("fail to convert end datetime: %w", err)
			}

			res = append(res, entity.Event{
				ID:         valueobject.EventID(item.Id),
				CalendarID: calendarID,
				Summary:    item.Summary,
				Start:      start,
				End:        end,
				Status:     item.Status,
			})
		}

		pageToken = events.NextPageToken
		syncToken = events.NextSyncToken

		logger.Debugf(ctx, "list events: pageToken=%q, syncToken=%q", pageToken, syncToken)
	}

	return res, syncToken, nil
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

func watch(ctx context.Context, service *calendar.Service, webhookBaseURL string, clockService *service.Clock,
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
