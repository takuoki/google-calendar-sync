package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
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

	syncToken, err := u.databaseRepo.GetLatestSyncToken(ctx, calendarID)
	if err != nil {
		return fmt.Errorf("fail to get latest sync token: %w", err)
	}

	var events []entity.Event
	var nextSyncToken string
	if syncToken != "" {
		// TODO: 定期的な予定の実装は未実施
		events, _, nextSyncToken, err = u.googleCalenderRepo.ListEventsWithSyncToken(ctx, calendarID, syncToken)
		if err != nil {
			if err == domain.SyncTokenIsOldError {
				// syncToken が古い場合は、全件取得して更新する
				u.logger.Info(ctx, "sync token is old, sync all events")
				events, nextSyncToken, err = u.syncAll(ctx, calendarID)
				if err != nil {
					return fmt.Errorf("fail to sync all events (sync token is old): %w", err)
				}
			} else {
				return fmt.Errorf("fail to list events: %w", err)
			}
		}
	} else {
		u.logger.Info(ctx, "sync all events")
		events, nextSyncToken, err = u.syncAll(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("fail to sync all events (sync token doesn't exist): %w", err)
		}
	}

	syncTime := u.clockService.Now()

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		if err := tx.LockCalendar(ctx, calendarID); err != nil {
			return fmt.Errorf("fail to lock calendar: %w", err)
		}

		updatedEventCount, err := tx.SyncEvents(ctx, calendarID, events)
		if err != nil {
			return fmt.Errorf("fail to sync events: %w", err)
		}

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

func (u *syncUsecase) syncAll(ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Event, string, error) {
	oneWeekAgo := u.clockService.Today().Add(7 * -24 * time.Hour)

	// TODO: 定期的な予定の実装は未実施
	events, _, nextSyncToken, err := u.googleCalenderRepo.ListEventsWithAfter(ctx, calendarID, oneWeekAgo)
	if err != nil {
		return nil, "", fmt.Errorf("fail to list events: %w", err)
	}

	return events, nextSyncToken, nil
}
