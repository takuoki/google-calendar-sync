package usecase

import (
	"context"
	"fmt"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

type WatchUsecase interface {
	StartAll(ctx context.Context) error
	Start(ctx context.Context, calendarID valueobject.CalendarID) error
	Stop(ctx context.Context, calendarID valueobject.CalendarID) error
}

type watchUsecase struct {
	googleCalenderRepo repository.GoogleCalendarRepository
	databaseRepo       repository.DatabaseRepository
	logger             applog.Logger
}

func NewWatchUsecase(
	googleCalenderRepo repository.GoogleCalendarRepository,
	databaseRepo repository.DatabaseRepository,
	logger applog.Logger,
) WatchUsecase {
	return &watchUsecase{
		googleCalenderRepo: googleCalenderRepo,
		databaseRepo:       databaseRepo,
		logger:             logger,
	}
}

func (u *watchUsecase) StartAll(ctx context.Context) error {

	calendars, err := u.databaseRepo.ListCalendars(ctx)
	if err != nil {
		return fmt.Errorf("fail to list calendars: %w", err)
	}

	for _, calendar := range calendars {
		if err := u.Start(ctx, calendar.ID); err != nil {
			return fmt.Errorf("fail to start (calendarID: %q): %w", calendar.ID, err)
		}
	}

	return nil
}

func (u *watchUsecase) Start(ctx context.Context, calendarID valueobject.CalendarID) error {

	if _, err := u.databaseRepo.GetCalendar(ctx, calendarID); err != nil {
		return fmt.Errorf("fail to get calendar: %w", err)
	}

	err := u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		if err := u.stopIfExistActiveChannel(ctx, tx, calendarID); err != nil {
			return fmt.Errorf("fail to stop: %w", err)
		}

		channel, err := u.googleCalenderRepo.Watch(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("fail to watch calendar: %w", err)
		}
		if channel == nil {
			return fmt.Errorf("fail to watch calendar: channel is nil")
		}

		if err := tx.CreateChannelHistory(ctx, *channel); err != nil {
			return fmt.Errorf("fail to create channel history: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}

func (u *watchUsecase) Stop(ctx context.Context, calendarID valueobject.CalendarID) error {

	if _, err := u.databaseRepo.GetCalendar(ctx, calendarID); err != nil {
		return fmt.Errorf("fail to get calendar: %w", err)
	}

	err := u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {
		if err := u.stopIfExistActiveChannel(ctx, tx, calendarID); err != nil {
			return fmt.Errorf("fail to stop: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}

func (u *watchUsecase) stopIfExistActiveChannel(
	ctx context.Context, tx repository.DatabaseTransaction, calendarID valueobject.CalendarID) error {

	// 複数チャネル取得できる形としているが、
	// アプリケーションの仕様上、取得されるのは最大 1 レコードのみとなる
	channels, err := tx.ListActiveChannelHistoriesWithLock(ctx, calendarID)
	if err != nil {
		return fmt.Errorf("fail to check active channel: %w", err)
	}

	if len(channels) == 0 {
		return nil
	}

	// API のみ成功し DB 更新が失敗した場合にステータス不整合が発生するが、
	// いずれ有効期限が切れるものなので許容する
	for _, channel := range channels {
		if err := u.googleCalenderRepo.StopWatch(ctx, channel); err != nil {
			return fmt.Errorf("fail to stop watch: %w", err)
		}
	}

	// SQL としては複数レコード更新される可能性があるが、
	// アプリケーションの仕様上、更新されるのは最大 1 レコードのみとなる
	if err := tx.StopActiveChannels(ctx, calendarID); err != nil {
		return fmt.Errorf("fail to stop active channels: %w", err)
	}

	return nil
}
