package usecase_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/takuoki/golib/applog"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func setupWatchUsecase(mockRepo repository.GoogleCalendarRepository) (usecase.WatchUsecase, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	logger, err := applog.NewSimpleLogger(buf)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	watchUsecase := usecase.NewWatchUsecase(mockRepo, mysqlRepo, logger)

	return watchUsecase, buf
}

func TestWatchUsecase_StartAll_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Given
	mockRepo := &GoogleCalendarRepositoryMock{
		WatchFunc: func(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {
			now := time.Now()
			return &entity.Channel{
				CalendarID: calendarID,
				ResourceID: "resource-id",
				StartTime:  now,
				Expiration: now.Add(1 * time.Hour),
			}, nil
		},
	}

	watchUsecase, _ := setupWatchUsecase(mockRepo)

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   "watch-all-success-1",
		Name: "Test Calendar 1",
	}))

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   "watch-all-success-2",
		Name: "Test Calendar 2",
	}))

	// When
	err := watchUsecase.StartAll(ctx)
	require.NoError(t, err)

	// Then
	_, err = mysqlRepo.GetChannelHistory(ctx, t, "watch-all-success-1")
	require.NoError(t, err)

	_, err = mysqlRepo.GetChannelHistory(ctx, t, "watch-all-success-2")
	require.NoError(t, err)
}
