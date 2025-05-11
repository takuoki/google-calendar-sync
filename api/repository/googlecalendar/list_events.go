package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

type listEventCall interface {
	Do(...googleapi.CallOption) (*calendar.Events, error)
	PageToken(string) listEventCall
}

type eventsListCallWrapper struct {
	call *calendar.EventsListCall
}

func (w *eventsListCallWrapper) Do(opts ...googleapi.CallOption) (*calendar.Events, error) {
	return w.call.Do(opts...)
}

func (w *eventsListCallWrapper) PageToken(token string) listEventCall {
	w.call = w.call.PageToken(token)
	return w
}

type eventsInstancesCallWrapper struct {
	call *calendar.EventsInstancesCall
}

func (w *eventsInstancesCallWrapper) Do(opts ...googleapi.CallOption) (*calendar.Events, error) {
	return w.call.Do(opts...)
}

func (w *eventsInstancesCallWrapper) PageToken(token string) listEventCall {
	w.call = w.call.PageToken(token)
	return w
}

func listEvents(ctx context.Context, clockService service.Clock, logger applog.Logger,
	baseCall listEventCall, calendarID valueobject.CalendarID) ([]entity.Event, string, error) {

	pageToken := ""
	syncToken := ""
	res := []entity.Event{}
	for syncToken == "" { // 最後のページまで取得すると必ず値が入る
		if err := ctx.Err(); err != nil {
			return nil, "", fmt.Errorf("context error: %w", err)
		}

		var call listEventCall
		if pageToken == "" {
			call = baseCall
		} else {
			call = baseCall.PageToken(pageToken)
		}

		events, err := call.Do()
		if err != nil {
			// WARNING: syncToken が古い場合については動作確認未実施
			// see: https://pkg.go.dev/google.golang.org/api/calendar/v3#EventsListCall.SyncToken
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 410 {
				return nil, "", domain.SyncTokenIsOldError
			}
			return nil, "", fmt.Errorf("fail to list events: %w", err)
		}

		for _, item := range events.Items {

			// TODO: この Debug ログは最終的には削除する（ログの量が多いため）
			itemJSON, err := json.Marshal(item)
			if err != nil {
				return nil, "", fmt.Errorf("fail to marshal event item to JSON: %w", err)
			}
			logger.Debugf(ctx, "event detail: %s", string(itemJSON))

			start, err := convertDateTime(item.Start, events.TimeZone)
			if err != nil {
				return nil, "", fmt.Errorf("fail to convert start datetime: %w", err)
			}
			end, err := convertDateTime(item.End, events.TimeZone)
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
