package usecase

import (
	"context"
	"fmt"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

type WatchUsecase interface {
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

func (u *watchUsecase) Start(ctx context.Context, calendarID valueobject.CalendarID) error {

	if _, err := u.databaseRepo.GetCalendar(ctx, calendarID); err != nil {
		return fmt.Errorf("fail to get calendar: %w", err)
	}

	channel, err := u.googleCalenderRepo.Watch(ctx, calendarID)
	if err != nil {
		return fmt.Errorf("fail to watch calendar: %w", err)
	}

	err = u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {

		// TODO: 既に登録されている Channel が存在する場合は停止させたい

		if err := tx.CreateChannel(ctx, *channel); err != nil {
			return fmt.Errorf("fail to create channel: %w", err)
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

		channels, err := tx.ListActiveChannels(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("fail to list active channels: %w", err)
		}

		for _, channel := range channels {
			// API のみ成功し DB 更新が失敗した場合にステータス不整合が発生するが、
			// いずれ有効期限が切れるものなので許容する
			if err := u.googleCalenderRepo.StopWatch(ctx, channel.ID); err != nil {
				return fmt.Errorf("fail to stop watch: %w", err)
			}

			if err := tx.StopChannel(ctx, channel.ID); err != nil {
				return fmt.Errorf("fail to stop channel: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}
