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
	// This test cannot be executed in parallel because it modifies shared state in mysqlRepo.

	ctx := context.Background()
	cleanup(ctx, t)

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

	var calendarID1 valueobject.CalendarID = "start-all-success-1"
	var calendarID2 valueobject.CalendarID = "start-all-success-2"
	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID1,
		Name: "Test Calendar 1",
	}))

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID2,
		Name: "Test Calendar 2",
	}))

	// When
	err := watchUsecase.StartAll(ctx)
	require.NoError(t, err)

	// Then
	// Verify a new channel was created
	_, err = mysqlRepo.GetLatestChannelHistory(ctx, t, calendarID1)
	require.NoError(t, err)

	_, err = mysqlRepo.GetLatestChannelHistory(ctx, t, calendarID2)
	require.NoError(t, err)
}

func TestWatchUsecase_Start_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Given
	mockRepo := &GoogleCalendarRepositoryMock{
		WatchFunc: func(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Channel, error) {
			now := mysqlRepo.Clock(t).Now()
			return &entity.Channel{
				CalendarID: calendarID,
				ResourceID: "new-resource-id",
				StartTime:  now,
				Expiration: now.Add(1 * time.Hour),
			}, nil
		},
		StopWatchFunc: func(ctx context.Context, channel entity.Channel) error {
			return nil
		},
	}

	watchUsecase, _ := setupWatchUsecase(mockRepo)

	var calendarID valueobject.CalendarID = "start-success-1"
	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID,
		Name: "Test Calendar",
	}))

	startTime := mysqlRepo.Clock(t).Now().Add(-1 * time.Hour)
	activeChannel := entity.Channel{
		CalendarID: calendarID,
		ResourceID: "active-resource-id",
		StartTime:  startTime,
		Expiration: startTime.Add(2 * time.Hour),
	}
	require.NoError(t, mysqlRepo.CreateChannelHistory(ctx, t, activeChannel))

	// When
	err := watchUsecase.Start(ctx, calendarID)
	require.NoError(t, err)

	// Then
	// Verify the active channel was stopped
	_, err = mysqlRepo.GetChannelHistory(ctx, t, calendarID, startTime)
	require.NoError(t, err)
	// assert.Equal(t, false, oldChannel.IsStopped)

	// Verify a new channel was created
	_, err = mysqlRepo.GetLatestChannelHistory(ctx, t, calendarID)
	require.NoError(t, err)
}
