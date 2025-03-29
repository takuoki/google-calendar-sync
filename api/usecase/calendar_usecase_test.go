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
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func setUpCalendarUsecase(useOauth bool) (usecase.CalendarUsecase, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	logger, err := applog.NewSimpleLogger(buf)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	calendarUsecase := usecase.NewCalendarUsecase(databaseRepo, useOauth, logger)

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
			calendarID:   valueobject.CalendarID("success-1"),
			name:         "Test Calendar 1",
			refreshToken: func() *string { s := "test-refresh-token"; return &s }(),
		},
		"without refresh token and useOauth false": {
			useOauth:     false,
			calendarID:   valueobject.CalendarID("success-2"),
			name:         "Test Calendar 2",
			refreshToken: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			calendarUsecase, _ := setUpCalendarUsecase(tt.useOauth)
			ctx := context.Background()

			err := calendarUsecase.Create(ctx, tt.calendarID, tt.name, tt.refreshToken)
			require.NoError(t, err)

			calendar, err := databaseRepo.GetCalendar(ctx, tt.calendarID)
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
			calendarID:   valueobject.CalendarID("failure-1"),
			name:         "Test Calendar 1",
			refreshToken: nil,
			errPrefix:    "refreshToken is required",
		},
		"empty refresh token with useOauth true": {
			useOauth:     true,
			calendarID:   valueobject.CalendarID("failure-2"),
			name:         "Test Calendar 2",
			refreshToken: func() *string { s := ""; return &s }(),
			errPrefix:    "refreshToken is required",
		},
		"refresh token provided with useOauth false": {
			useOauth:     false,
			calendarID:   valueobject.CalendarID("failure-3"),
			name:         "Test Calendar 3",
			refreshToken: func() *string { s := "unexpected-token"; return &s }(),
			errPrefix:    "refreshToken is not allowed",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			calendarUsecase, _ := setUpCalendarUsecase(tt.useOauth)
			ctx := context.Background()

			err := calendarUsecase.Create(ctx, tt.calendarID, tt.name, tt.refreshToken)
			require.Error(t, err)

			if !strings.HasPrefix(err.Error(), tt.errPrefix) {
				t.Errorf("error message does not match the expected prefix, got: %s", err.Error())
			}
		})
	}
}

func TestCalendarUsecase_Create_DuplicateError(t *testing.T) {
	t.Parallel()

	calendarID := valueobject.CalendarID("duplicate-id")
	name := "Duplicate Calendar"

	// Set up initial calendar
	calendarUsecase, _ := setUpCalendarUsecase(false)
	ctx := context.Background()
	err := calendarUsecase.Create(ctx, calendarID, name, nil)
	require.NoError(t, err)

	// Attempt to create a calendar with the same ID
	err = calendarUsecase.Create(ctx, calendarID, name, nil)
	require.Error(t, err)

	if !strings.HasPrefix(err.Error(), "fail to run transaction") {
		t.Errorf("error message does not match the expected prefix, got: %s", err.Error())
	}
}
