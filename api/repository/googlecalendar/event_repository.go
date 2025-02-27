package googlecalendar

import (
	"context"
	"fmt"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *googleCalendarRepository) ListEvents(
	ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Event, string, error) {
	call := r.service.Events.List(string(calendarID)).Context(ctx)
	return r.listEvents(ctx, call, calendarID)
}

func (r *googleCalendarRepository) ListEventsWithSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, string, error) {
	call := r.service.Events.List(string(calendarID)).Context(ctx).SyncToken(syncToken)
	return r.listEvents(ctx, call, calendarID)
}

func (r *googleCalendarRepository) listEvents(
	ctx context.Context, baseCall *calendar.EventsListCall, calendarID valueobject.CalendarID) ([]entity.Event, string, error) {

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

			r.logger.Debugf(ctx, "list events: item=%+v", item)

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

		r.logger.Debugf(ctx, "list events: pageToken=%q, syncToken=%q", pageToken, syncToken)
	}

	return res, syncToken, nil
}

func (r *googleCalendarRepository) Watch(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {

	// TODO: ttl の設定を追加したい
	request := calendar.Channel{
		Id:      calendarID.ToChannelID(),
		Type:    "web_hook",
		Address: fmt.Sprintf("%s/%s/", r.webhookBaseURL, calendarID),
	}

	channel, err := r.service.Events.Watch(string(calendarID), &request).Context(ctx).Do()
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
		StartTime:  r.clockService.Now(),
		Expiration: expiration,
	}, nil
}
