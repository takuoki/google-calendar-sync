package usecase_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takuoki/golib/applog"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func setupCalendarUsecase(useOauth bool) (usecase.CalendarUsecase, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	logger, err := applog.NewSimpleLogger(buf)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	calendarUsecase := usecase.NewCalendarUsecase(mysqlRepo, useOauth, logger)

	return calendarUsecase, buf
}

func TestCalendarUsecase_Create_Success(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		useOauth     bool
		calendarID   valueobject.CalendarID
		name         string
		refreshToken *string
	}{
		"with refresh token and useOauth true": {
			useOauth:     true,
			calendarID:   "calendar-success-1",
			name:         "Test Calendar 1",
			refreshToken: func() *string { s := "test-refresh-token"; return &s }(),
		},
		"without refresh token and useOauth false": {
			useOauth:     false,
			calendarID:   "calendar-success-2",
			name:         "Test Calendar 2",
			refreshToken: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Given
			calendarUsecase, _ := setupCalendarUsecase(tt.useOauth)

			// When
			err := calendarUsecase.Create(ctx, tt.calendarID, tt.name, tt.refreshToken)
			require.NoError(t, err)

			// Then
			calendar, err := mysqlRepo.GetCalendar(ctx, tt.calendarID)
			require.NoError(t, err)

			assert.Equal(t, tt.calendarID, calendar.ID)
			assert.Equal(t, tt.name, calendar.Name)
			if tt.refreshToken != nil {
				assert.Equal(t, *tt.refreshToken, *calendar.RefreshToken)
			} else {
				assert.Nil(t, calendar.RefreshToken)
			}
		})
	}
}

func TestCalendarUsecase_Create_Failure(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		useOauth     bool
		calendarID   valueobject.CalendarID
		name         string
		refreshToken *string
		errPrefix    string
	}{
		"missing refresh token with useOauth true": {
			useOauth:     true,
			calendarID:   "calendar-failure-1",
			name:         "Test Calendar 1",
			refreshToken: nil,
			errPrefix:    "refreshToken is required",
		},
		"empty refresh token with useOauth true": {
			useOauth:     true,
			calendarID:   "calendar-failure-2",
			name:         "Test Calendar 2",
			refreshToken: func() *string { s := ""; return &s }(),
			errPrefix:    "refreshToken is required",
		},
		"refresh token provided with useOauth false": {
			useOauth:     false,
			calendarID:   "calendar-failure-3",
			name:         "Test Calendar 3",
			refreshToken: func() *string { s := "unexpected-token"; return &s }(),
			errPrefix:    "refreshToken is not allowed",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Given
			calendarUsecase, _ := setupCalendarUsecase(tt.useOauth)

			// When
			err := calendarUsecase.Create(ctx, tt.calendarID, tt.name, tt.refreshToken)
			require.Error(t, err)

			// Then
			if !strings.HasPrefix(err.Error(), tt.errPrefix) {
				t.Errorf("error message does not match the expected prefix, got: %s", err.Error())
			}
		})
	}
}

func TestCalendarUsecase_Create_DuplicateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Given
	calendarUsecase, _ := setupCalendarUsecase(false)

	var calendarID valueobject.CalendarID = "calendar-duplicate-id"
	name := "Duplicate Calendar"

	err := mysqlRepo.CreateCalendar(ctx, t, entity.Calendar{
		ID:   calendarID,
		Name: name,
	})
	require.NoError(t, err)

	// When
	err = calendarUsecase.Create(ctx, calendarID, name, nil)
	require.Error(t, err)

	// Then
	if !strings.HasPrefix(err.Error(), "fail to run transaction") {
		t.Errorf("error message does not match the expected prefix, got: %s", err.Error())
	}
}
