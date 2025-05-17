package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/constant"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

const (
	syncEventFrom         = -1 * 7 * 24 * time.Hour // 1 週間前
	syncEventInstanceFrom = syncEventFrom
	syncEventInstanceTo   = 365 * 24 * time.Hour // 1 年後
)

type SyncUsecase interface {
	Sync(ctx context.Context, calendarID valueobject.CalendarID) error
}

type syncUsecase struct {
	clockService       service.Clock
	googleCalenderRepo repository.GoogleCalendarRepository
	databaseRepo       repository.DatabaseRepository
	logger             applog.Logger
}

func NewSyncUsecase(
	clockService service.Clock,
	googleCalenderRepo repository.GoogleCalendarRepository,
	databaseRepo repository.DatabaseRepository,
	logger applog.Logger,
) SyncUsecase {
	return &syncUsecase{
		clockService:       clockService,
		googleCalenderRepo: googleCalenderRepo,
		databaseRepo:       databaseRepo,
		logger:             logger,
	}
}

func (u *syncUsecase) Sync(ctx context.Context, calendarID valueobject.CalendarID) error {

	if _, err := u.databaseRepo.GetCalendar(ctx, calendarID); err != nil {
		return fmt.Errorf("fail to get calendar: %w", err)
	}

	events, recurringEvents, nextSyncToken, err := u.listEventsFromGoogleCalendar(ctx, calendarID)
	if err != nil {
		return fmt.Errorf("fail to list events from Google Calendar: %w", err)
	}

	syncTime := u.clockService.Now()

	shouldSaveRecurringEvents, eventInstanceMap, err := u.listEventInstancesFromGoogleCalendar(
		ctx, recurringEvents, calendarID, syncTime)
	if err != nil {
		return fmt.Errorf("fail to list event instances from Google Calendar: %w", err)
	}

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		if err := tx.LockCalendar(ctx, calendarID); err != nil {
			return fmt.Errorf("fail to lock calendar: %w", err)
		}

		updatedEventCount := 0

		// events には定期イベントの個別イベントが含まれる可能性があるため、先に定期イベントを登録する
		for _, recurringEvent := range shouldSaveRecurringEvents {
			instances := eventInstanceMap[recurringEvent.ID]
			cnt, err := tx.SyncRecurringEventAndInstancesWithAfter(
				ctx, recurringEvent, instances, syncTime.Add(syncEventInstanceFrom))
			if err != nil {
				return fmt.Errorf("fail to sync recurring events: %w", err)
			}

			updatedEventCount += cnt
		}

		cnt, err := tx.SyncEvents(ctx, calendarID, events)
		if err != nil {
			return fmt.Errorf("fail to sync events: %w", err)
		}

		updatedEventCount += cnt

		if err := tx.CreateSyncHistory(ctx, calendarID, syncTime, nextSyncToken, updatedEventCount); err != nil {
			return fmt.Errorf("fail to create sync history: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}

func (u *syncUsecase) listEventsFromGoogleCalendar(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Event, []entity.RecurringEvent, string, error) {
	syncToken, err := u.databaseRepo.GetLatestSyncToken(ctx, calendarID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fail to get latest sync token: %w", err)
	}

	var events []entity.Event
	var recurringEvents []entity.RecurringEvent
	var nextSyncToken string
	if syncToken != "" {
		events, recurringEvents, nextSyncToken, err = u.googleCalenderRepo.ListEventsWithSyncToken(ctx, calendarID, syncToken)
		if err != nil {
			if err == domain.SyncTokenIsOldError {
				// syncToken が古い場合は、全件取得して更新する
				u.logger.Info(ctx, "sync token is old, sync all events")
				events, recurringEvents, nextSyncToken, err = u.listAllEventsFromGoogleCalendar(ctx, calendarID)
				if err != nil {
					return nil, nil, "", fmt.Errorf("fail to sync all events (sync token is old): %w", err)
				}
			} else {
				return nil, nil, "", fmt.Errorf("fail to list events: %w", err)
			}
		}
	} else {
		u.logger.Info(ctx, "sync all events")
		events, recurringEvents, nextSyncToken, err = u.listAllEventsFromGoogleCalendar(ctx, calendarID)
		if err != nil {
			return nil, nil, "", fmt.Errorf("fail to sync all events (sync token doesn't exist): %w", err)
		}
	}

	return events, recurringEvents, nextSyncToken, nil
}

func (u *syncUsecase) listAllEventsFromGoogleCalendar(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Event, []entity.RecurringEvent, string, error) {
	after := u.clockService.Today().Add(syncEventFrom)

	events, recurringEvents, nextSyncToken, err := u.googleCalenderRepo.ListEventsWithAfter(ctx, calendarID, after)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fail to list events: %w", err)
	}

	return events, recurringEvents, nextSyncToken, nil
}

func (u *syncUsecase) listEventInstancesFromGoogleCalendar(ctx context.Context,
	recurringEvents []entity.RecurringEvent, calendarID valueobject.CalendarID, syncTime time.Time) (
	[]entity.RecurringEvent, map[valueobject.EventID][]entity.Event, error) {

	if len(recurringEvents) == 0 {
		return nil, nil, nil
	}

	shouldSaveRecurringEvents := make([]entity.RecurringEvent, 0, len(recurringEvents))
	eventInstanceMap := map[valueobject.EventID][]entity.Event{}

	// ここでは終了日が到達していない定期イベントのみを取得する
	// 終了日が到達した定期イベントの終了日が延期された場合はここでは取得されず、
	// 新規定期イベントと同様の挙動となり、後続の SyncRecurringEventAndInstancesWithAfter が呼ばれる
	// （登録時に再度、存在チェックを行なっているため、新規登録ではなく更新処理となる）
	recurringEventMap, err := u.fetchRecurringEventMapWithAfter(ctx, calendarID, syncTime.Add(syncEventFrom))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to fetch recurring event map: %w", err)
	}

	for _, recurringEvent := range recurringEvents {
		recurringEvent, ok := recurringEventMap[recurringEvent.ID]
		if ok && recurringEvent.Equals(&recurringEvent) {
			// 既存の定期イベントと同じ場合はスキップ
			continue
		}

		var instances []entity.Event
		if recurringEvent.Status != constant.EventStatusCancelled {
			instances, err = u.googleCalenderRepo.ListEventInstancesBetween(
				ctx, calendarID, recurringEvent.ID,
				syncTime.Add(syncEventInstanceFrom), syncTime.Add(syncEventInstanceTo))
			if err != nil {
				return nil, nil, fmt.Errorf("fail to list event instances: %w", err)
			}
		}

		shouldSaveRecurringEvents = append(shouldSaveRecurringEvents, recurringEvent)
		eventInstanceMap[recurringEvent.ID] = instances
	}

	return shouldSaveRecurringEvents, eventInstanceMap, nil
}

func (u *syncUsecase) fetchRecurringEventMapWithAfter(ctx context.Context, calendarID valueobject.CalendarID, after time.Time) (
	map[valueobject.EventID]entity.RecurringEvent, error) {

	recurringEvents, err := u.databaseRepo.ListActiveRecurringEventsWithAfter(ctx, calendarID, after)
	if err != nil {
		return nil, fmt.Errorf("fail to list recurring events: %w", err)
	}

	recurringEventMap := map[valueobject.EventID]entity.RecurringEvent{}
	for _, recurringEvent := range recurringEvents {
		recurringEventMap[recurringEvent.ID] = recurringEvent
	}

	return recurringEventMap, nil
}
