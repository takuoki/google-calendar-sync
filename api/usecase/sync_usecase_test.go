package usecase_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func setupSyncUsecase(clockService service.Clock, mockRepo repository.GoogleCalendarRepository) (usecase.SyncUsecase, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	logger, err := applog.NewSimpleLogger(buf)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	syncUsecase := usecase.NewSyncUsecase(clockService, mockRepo, mysqlRepo, logger)

	return syncUsecase, buf
}

func TestSyncUsecase_Sync_Success_AllEvents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClock := service.NewMockClock()

	p := func(t time.Time) *time.Time {
		return &t
	}

	// Given
	var calendarID valueobject.CalendarID = "sync-success-all-events-1"

	event1 := entity.Event{
		ID:         "event-1",
		CalendarID: calendarID,
		Summary:    "Test Event 1",
		Start:      p(mockClock.Now().Add(12 * time.Hour)),
		End:        p(mockClock.Now().Add(13 * time.Hour)),
		Status:     "confirmed",
	}
	event2 := entity.Event{
		ID:         "event-2",
		CalendarID: calendarID,
		Summary:    "Test Event 2",
		Start:      p(mockClock.Now().Add(23 * time.Hour)),
		End:        p(mockClock.Now().Add(24 * time.Hour)),
		Status:     "confirmed",
	}

	mockRepo := &GoogleCalendarRepositoryMock{
		ListEventsWithAfterFunc: func(ctx context.Context,
			calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, []entity.RecurringEvent, string, error) {
			return []entity.Event{event1, event2}, []entity.RecurringEvent{}, "new-sync-token", nil
		},
	}

	syncUsecase, buf := setupSyncUsecase(mockClock, mockRepo)

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID,
		Name: "Test Calendar",
	}))

	// When
	err := syncUsecase.Sync(ctx, calendarID)
	require.NoError(t, err)

	// Then
	// Verify a new sync history was created
	syncToken, err := mysqlRepo.GetLatestSyncToken(ctx, calendarID)
	require.NoError(t, err)
	assert.Equal(t, "new-sync-token", syncToken)

	// Verify some events were created
	events, err := mysqlRepo.ListEvents(ctx, t, calendarID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assertEqualEvent(t, event1, events[0])
	assertEqualEvent(t, event2, events[1])

	// Verify log messages
	logs := strings.Split(buf.String(), "\n")
	require.Contains(t, logs, "sync all events")
}

func TestSyncUsecase_Sync_Success_WithSyncToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClock := service.NewMockClock()

	p := func(t time.Time) *time.Time {
		return &t
	}

	// Given
	var calendarID valueobject.CalendarID = "sync-success-with-sync-token-1"

	event1 := entity.Event{
		ID:         "event-1",
		CalendarID: calendarID,
		Summary:    "",
		Start:      nil,
		End:        nil,
		Status:     "cancelled",
	}
	event2 := entity.Event{
		ID:         "event-2",
		CalendarID: calendarID,
		Summary:    "Test Event 2",
		Start:      p(mockClock.Now().Add(12 * time.Hour)),
		End:        p(mockClock.Now().Add(13 * time.Hour)),
		Status:     "confirmed",
	}

	mockRepo := &GoogleCalendarRepositoryMock{
		ListEventsWithSyncTokenFunc: func(ctx context.Context,
			calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, []entity.RecurringEvent, string, error) {
			assert.Equal(t, "sync-token", syncToken)
			return []entity.Event{event1, event2}, []entity.RecurringEvent{}, "new-sync-token", nil
		},
	}

	syncUsecase, buf := setupSyncUsecase(mockClock, mockRepo)

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID,
		Name: "Test Calendar",
	}))

	require.NoError(t, mysqlRepo.CreateEvent(ctx, t, entity.Event{
		ID:         event1.ID,
		CalendarID: calendarID,
		Summary:    "Test Event 1",
		Start:      p(mockClock.Now().Add(2 * time.Hour)),
		End:        p(mockClock.Now().Add(3 * time.Hour)),
		Status:     "confirmed",
	}))

	require.NoError(t, mysqlRepo.CreateSyncHistory(ctx, t,
		calendarID, mockClock.Now().Add(-1*time.Hour), "sync-token", 0))

	// When
	err := syncUsecase.Sync(ctx, calendarID)
	require.NoError(t, err)

	// Then
	// Verify a new sync history was created
	syncToken, err := mysqlRepo.GetLatestSyncToken(ctx, calendarID)
	require.NoError(t, err)
	assert.Equal(t, "new-sync-token", syncToken)

	// Verify some events were created
	events, err := mysqlRepo.ListEvents(ctx, t, calendarID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assertEqualEvent(t, event1, events[0])
	assertEqualEvent(t, event2, events[1])

	// Verify log messages
	assert.Equal(t, "", buf.String())
}

func TestSyncUsecase_Sync_Success_OldSyncToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClock := service.NewMockClock()

	p := func(t time.Time) *time.Time {
		return &t
	}

	// Given
	var calendarID valueobject.CalendarID = "sync-success-old-sync-token-1"

	event1 := entity.Event{
		ID:         "event-1",
		CalendarID: calendarID,
		Summary:    "",
		Start:      nil,
		End:        nil,
		Status:     "cancelled",
	}
	event2 := entity.Event{
		ID:         "event-2",
		CalendarID: calendarID,
		Summary:    "Test Event 2",
		Start:      p(mockClock.Now().Add(12 * time.Hour)),
		End:        p(mockClock.Now().Add(13 * time.Hour)),
		Status:     "confirmed",
	}

	mockRepo := &GoogleCalendarRepositoryMock{
		ListEventsWithSyncTokenFunc: func(ctx context.Context,
			calendarID valueobject.CalendarID, syncToken string) ([]entity.Event, []entity.RecurringEvent, string, error) {
			return nil, nil, "", domain.SyncTokenIsOldError
		},
		ListEventsWithAfterFunc: func(ctx context.Context,
			calendarID valueobject.CalendarID, after time.Time) ([]entity.Event, []entity.RecurringEvent, string, error) {
			return []entity.Event{event1, event2}, []entity.RecurringEvent{}, "new-sync-token", nil
		},
	}

	syncUsecase, buf := setupSyncUsecase(mockClock, mockRepo)

	require.NoError(t, mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID,
		Name: "Test Calendar",
	}))

	require.NoError(t, mysqlRepo.CreateEvent(ctx, t, entity.Event{
		ID:         event1.ID,
		CalendarID: calendarID,
		Summary:    "Test Event 1",
		Start:      p(mockClock.Now().Add(2 * time.Hour)),
		End:        p(mockClock.Now().Add(3 * time.Hour)),
		Status:     "confirmed",
	}))

	require.NoError(t, mysqlRepo.CreateSyncHistory(ctx, t,
		calendarID, mockClock.Now().Add(-1*time.Hour), "sync-token", 0))

	// When
	err := syncUsecase.Sync(ctx, calendarID)
	require.NoError(t, err)

	// Then
	// Verify a new sync history was created
	syncToken, err := mysqlRepo.GetLatestSyncToken(ctx, calendarID)
	require.NoError(t, err)
	assert.Equal(t, "new-sync-token", syncToken)

	// Verify some events were created
	events, err := mysqlRepo.ListEvents(ctx, t, calendarID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assertEqualEvent(t, event1, events[0])
	assertEqualEvent(t, event2, events[1])

	// Verify log messages
	logs := strings.Split(buf.String(), "\n")
	require.Contains(t, logs, "sync token is old, sync all events")
}
