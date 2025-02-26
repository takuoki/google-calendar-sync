package usecase

import (
	"context"
	"fmt"

	"github.com/takuoki/golib/applog"
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
	if syncToken == "" {
		// TODO: 過去分すべてをSyncするのは大変なので、本日以降のものだけにしたい
		u.logger.Info(ctx, "sync all events")
		events, nextSyncToken, err = u.googleCalenderRepo.ListEvents(ctx, calendarID)
	} else {
		events, nextSyncToken, err = u.googleCalenderRepo.ListEventsWithSyncToken(ctx, calendarID, syncToken)
	}
	if err != nil {
		return fmt.Errorf("fail to list events: %w", err)
	}

	syncTime := u.clockService.Now()

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {
		updatedEventCount, err := tx.SyncEvents(ctx, events)
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
