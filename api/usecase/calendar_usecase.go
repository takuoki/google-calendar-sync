package usecase

import (
	"context"
	"fmt"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

type CalendarUsecase interface {
	Create(ctx context.Context, calendarID valueobject.CalendarID, name string, refreshToken *string) error
}

type calendarUsecase struct {
	databaseRepo repository.DatabaseRepository
	useOauth     bool
	logger       applog.Logger
}

func NewCalendarUsecase(
	databaseRepo repository.DatabaseRepository,
	useOauth bool,
	logger applog.Logger,
) CalendarUsecase {
	return &calendarUsecase{
		databaseRepo: databaseRepo,
		useOauth:     useOauth,
		logger:       logger,
	}
}

func (u *calendarUsecase) Create(ctx context.Context, calendarID valueobject.CalendarID,
	name string, refreshToken *string) error {

	if u.useOauth && refreshToken == nil {
		// TODO: ClientError
		return fmt.Errorf("refresh token is required")
	}

	if !u.useOauth && refreshToken != nil {
		// TODO: ClientError
		return fmt.Errorf("refresh token is not required")
	}

	err := u.databaseRepo.RunTransaction(ctx, func(ctx context.Context, tx repository.DatabaseTransaction) error {
		calendar := entity.Calendar{
			ID:           calendarID,
			Name:         name,
			RefreshToken: refreshToken,
		}
		if err := tx.CreateCalendar(ctx, calendar); err != nil {
			return fmt.Errorf("fail to create calendar: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("fail to run transaction: %w", err)
	}

	return nil
}
