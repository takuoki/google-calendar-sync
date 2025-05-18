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
	syncEventFrom              = -1 * 7 * 24 * time.Hour // 1 週間前
	syncEventInstanceFrom      = syncEventFrom
	syncEventInstanceTo        = 365 * 24 * time.Hour     // 1 年後
	syncFutureInstanceInterval = (7 + 1) * 24 * time.Hour // 1 週間 + バッファ
)

type SyncUsecase interface {
	Sync(ctx context.Context, calendarID valueobject.CalendarID) error
	SyncFutureInstanceAll(ctx context.Context) error
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

	// 定期イベントの削除や Recurrence の削除を考慮し、 events と recurringEvents を更新する
	events, recurringEvents, err = u.moveOrCopyCancelledRecurringEvents(ctx, events, recurringEvents)
	if err != nil {
		return fmt.Errorf("fail to move cancelled recurring events: %w", err)
	}

	shouldSaveRecurringEvents, eventInstanceMap, err := u.listEventInstancesFromGoogleCalendar(
		ctx, calendarID, recurringEvents, syncTime)
	if err != nil {
		return fmt.Errorf("fail to list event instances from Google Calendar: %w", err)
	}

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		u.logger.Trace(ctx, "start transaction")

		if err := tx.LockCalendar(ctx, calendarID); err != nil {
			return fmt.Errorf("fail to lock calendar: %w", err)
		}

		updatedEventCount := 0

		// events には定期イベントの個別イベントが含まれる可能性があるため、先に定期イベントを登録する
		for _, recurringEvent := range shouldSaveRecurringEvents {
			u.logger.Tracef(ctx, "sync recurring event: eventID=%q", recurringEvent.ID)

			instances := eventInstanceMap[recurringEvent.ID]
			cnt, err := tx.SyncRecurringEventAndInstancesWithAfter(
				ctx, recurringEvent, instances, syncTime.Add(syncEventInstanceFrom))
			if err != nil {
				return fmt.Errorf("fail to sync recurring events: %w", err)
			}

			updatedEventCount += cnt
		}

		u.logger.Trace(ctx, "sync events")
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

// listEventsFromGoogleCalendar retrieves events and recurring events from a Google Calendar.
//
// It first attempts to fetch events using the latest sync token stored in the database.
// If the sync token is outdated or does not exist, it falls back to fetching all events from the calendar.
//
// Recurring event and non-recurring event are determined by the presence or absence
// of the latest recurring setting (`recurrence`).
// Note that if a recurring event is changed to a non-recurring event,
// the recurring event before the change is not included (and vice versa).
//
// Returns:
// - events: A slice of non-recurring events retrieved from the calendar.
// - recurringEvents: A slice of recurring events retrieved from the calendar.
// - nextSyncToken: A string representing the next sync token to be used for subsequent sync operations.
// - error: An error object if the operation fails, or nil if successful.
func (u *syncUsecase) listEventsFromGoogleCalendar(ctx context.Context, calendarID valueobject.CalendarID) (
	[]entity.Event, []entity.RecurringEvent, string, error) {
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

func (u *syncUsecase) listAllEventsFromGoogleCalendar(ctx context.Context, calendarID valueobject.CalendarID) (
	[]entity.Event, []entity.RecurringEvent, string, error) {
	after := u.clockService.Today().Add(syncEventFrom)

	events, recurringEvents, nextSyncToken, err := u.googleCalenderRepo.ListEventsWithAfter(ctx, calendarID, after)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fail to list events: %w", err)
	}

	return events, recurringEvents, nextSyncToken, nil
}

func (u *syncUsecase) moveOrCopyCancelledRecurringEvents(ctx context.Context,
	events []entity.Event, recurringEvents []entity.RecurringEvent) (
	[]entity.Event, []entity.RecurringEvent, error) {

	if len(events) == 0 {
		return events, recurringEvents, nil
	}

	eventIDs := make([]valueobject.EventID, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID)
	}

	// 定期イベントテーブルに有効で存在するものを取得
	// すでにキャンセルされている場合は、子イベントもキャンセルになっているはずなので抽出対象外
	dbRecurringEvents, err := u.databaseRepo.ListActiveRecurringEventsWithIDs(ctx, events[0].CalendarID, eventIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to list recurring events: %w", err)
	}
	recurringEventMap, err := u.convertToRecurringEventMap(dbRecurringEvents)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to convert recurring events: %w", err)
	}

	resEvents := make([]entity.Event, 0, len(events))
	resRecurringEvents := recurringEvents
	for _, event := range events {
		if _, ok := recurringEventMap[event.ID]; !ok {
			// DB に存在しない定期イベントは通常のイベントとして扱う
			resEvents = append(resEvents, event)
			continue
		}

		cancelledRecurringEvent := entity.NewCancelledRecurringEventFromEvent(event)
		if event.Status == constant.EventStatusCancelled {
			// イベントのステータスがキャンセルの場合、通常のイベントとしては扱わず、定期イベントをキャンセル扱いにする
			resRecurringEvents = append(resRecurringEvents, cancelledRecurringEvent)
		} else {
			// イベントのステータスがキャンセル以外の場合、通常のイベントとして扱いつつ、定期イベントをキャンセル扱いにする
			resEvents = append(resEvents, event)
			resRecurringEvents = append(resRecurringEvents, cancelledRecurringEvent)
		}
	}

	return resEvents, resRecurringEvents, nil
}

func (u *syncUsecase) listEventInstancesFromGoogleCalendar(ctx context.Context,
	calendarID valueobject.CalendarID, recurringEvents []entity.RecurringEvent, syncTime time.Time) (
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
	dbRecurringEvents, err := u.databaseRepo.ListActiveRecurringEventsWithAfter(ctx, calendarID, syncTime.Add(syncEventFrom))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to list recurring events: %w", err)
	}
	recurringEventMap, err := u.convertToRecurringEventMap(dbRecurringEvents)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to convert recurring events: %w", err)
	}

	for _, recurringEvent := range recurringEvents {
		dbRecurringEvent, ok := recurringEventMap[recurringEvent.ID]
		if ok && recurringEvent.Equals(&dbRecurringEvent) {
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

func (u *syncUsecase) convertToRecurringEventMap(recurringEvents []entity.RecurringEvent) (
	map[valueobject.EventID]entity.RecurringEvent, error) {

	recurringEventMap := map[valueobject.EventID]entity.RecurringEvent{}
	for _, recurringEvent := range recurringEvents {
		recurringEventMap[recurringEvent.ID] = recurringEvent
	}

	return recurringEventMap, nil
}

func (u *syncUsecase) SyncFutureInstanceAll(ctx context.Context) error {
	calendars, err := u.databaseRepo.ListCalendars(ctx)
	if err != nil {
		return fmt.Errorf("fail to list calendars: %w", err)
	}

	now := u.clockService.Now()

	for _, calendar := range calendars {
		if err := u.syncFutureInstance(ctx, calendar.ID, now); err != nil {
			return fmt.Errorf("fail to sync future instance (calendarID: %q): %w", calendar.ID, err)
		}
	}

	return nil
}

func (u *syncUsecase) syncFutureInstance(ctx context.Context, calendarID valueobject.CalendarID, baseTime time.Time) error {

	recurringEvents, err := u.databaseRepo.ListActiveRecurringEventsWithAfter(ctx, calendarID, baseTime.Add(syncEventFrom))
	if err != nil {
		return fmt.Errorf("fail to list recurring events: %w", err)
	}

	from := baseTime.Add(syncEventInstanceTo - syncFutureInstanceInterval)
	to := baseTime.Add(syncEventInstanceTo)

	shouldSaveRecurringEvents, eventInstanceMap, err := u.listFutureInstancesFromGoogleCalendar(ctx, calendarID, recurringEvents, from, to)
	if err != nil {
		return fmt.Errorf("fail to list future instances: %w", err)
	}

	if len(shouldSaveRecurringEvents) == 0 {
		return nil
	}

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		u.logger.Trace(ctx, "start transaction")

		if err := tx.LockCalendar(ctx, calendarID); err != nil {
			return fmt.Errorf("fail to lock calendar: %w", err)
		}

		updatedEventCount := 0

		for _, recurringEvent := range shouldSaveRecurringEvents {
			u.logger.Tracef(ctx, "sync recurring event: eventID=%q", recurringEvent.ID)

			instances := eventInstanceMap[recurringEvent.ID]
			// TODO: 生きているイベントが間違ってキャンセルされていないことを確認すべき
			cnt, err := tx.SyncRecurringEventAndInstancesWithAfter(
				ctx, recurringEvent, instances, from)
			if err != nil {
				return fmt.Errorf("fail to sync recurring events: %w", err)
			}

			updatedEventCount += cnt
		}

		// TODO: どのように記録を残すべきか要検討（syncToken があるわけではないので、必ずしも残さなくても良い）

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}

func (u *syncUsecase) listFutureInstancesFromGoogleCalendar(ctx context.Context,
	calendarID valueobject.CalendarID, recurringEvents []entity.RecurringEvent, from, to time.Time) (
	[]entity.RecurringEvent, map[valueobject.EventID][]entity.Event, error) {
	// listEventInstancesFromGoogleCalendar と異なり、引数の recurringEvents は DB から取得したもの
	// RecurringEvent 自体は更新されていないはずではあるが、 SyncRecurringEventAndInstancesWithAfter を利用したいため戻り値として返す

	if len(recurringEvents) == 0 {
		return nil, nil, nil
	}

	shouldSaveRecurringEvents := make([]entity.RecurringEvent, 0, len(recurringEvents))
	eventInstanceMap := map[valueobject.EventID][]entity.Event{}

	for _, recurringEvent := range recurringEvents {
		if recurringEvent.Status != constant.EventStatusCancelled {
			continue
		}

		instances, err := u.googleCalenderRepo.ListEventInstancesBetween(
			ctx, calendarID, recurringEvent.ID, from, to)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to list event instances: %w", err)
		}

		if len(instances) == 0 {
			continue
		}

		shouldSaveRecurringEvents = append(shouldSaveRecurringEvents, recurringEvent)
		eventInstanceMap[recurringEvent.ID] = instances
	}

	return shouldSaveRecurringEvents, eventInstanceMap, nil
}
