package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/repository/mysql"
)

var mysqlRepo *mysql.MysqlRepository

func TestMain(m *testing.M) {

	ctx := context.Background()

	db, err := mysql.ConnectDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		panic("fail to connect to the database: " + err.Error())
	}
	defer func() {
		cleanupForMain(ctx, m)
		if err := db.Close(); err != nil {
			panic("fail to close database: " + err.Error())
		}
	}()

	if err := waitForDatabaseReady(ctx, db); err != nil {
		panic("fail to wait for db ready: " + err.Error())
	}

	// Use io.Discard to ignore repository logs in usecase tests
	logger, err := applog.NewSimpleLogger(io.Discard)
	if err != nil {
		panic("fail to create logger: " + err.Error())
	}

	mysqlRepo = mysql.NewMysqlRepository(db, service.NewMockClock(), nil, logger)

	m.Run()
}

func waitForDatabaseReady(ctx context.Context, db *sql.DB) error {
	const (
		pingRetryLimit    = 10
		pingRetryInterval = 5 * time.Second
	)

	for i := 0; ; i++ {
		if i >= pingRetryLimit {
			return fmt.Errorf("fail to connect to the database after reaching the retry limit")
		}
		if err := db.PingContext(ctx); err != nil {
			time.Sleep(pingRetryInterval)
			continue
		}

		break
	}

	return nil
}

func cleanupForMain(ctx context.Context, m *testing.M) {
	if err := mysqlRepo.DeleteAllSyncHistoriesForMain(ctx, m); err != nil {
		panic("fail to delete all sync histories: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllChannelHistoriesForMain(ctx, m); err != nil {
		panic("fail to delete all channel histories: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllEventsForMain(ctx, m); err != nil {
		panic("fail to delete all events: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllCalendarsForMain(ctx, m); err != nil {
		panic("fail to delete all calendars: " + err.Error())
	}
}

func cleanup(ctx context.Context, t *testing.T) {
	t.Helper()

	if err := mysqlRepo.DeleteAllSyncHistories(ctx, t); err != nil {
		panic("fail to delete all sync histories: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllChannelHistories(ctx, t); err != nil {
		panic("fail to delete all channel histories: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllEvents(ctx, t); err != nil {
		panic("fail to delete all events: " + err.Error())
	}
	if err := mysqlRepo.DeleteAllCalendars(ctx, t); err != nil {
		panic("fail to delete all calendars: " + err.Error())
	}
}

func assertEqualTime(t *testing.T, expected, actual *time.Time) bool {
	t.Helper()

	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// ignore Timezone
	return expected.UTC().Equal(actual.UTC())
}

func assertEqualEvent(t *testing.T, expected, actual entity.Event) bool {
	t.Helper()

	if !assert.Equal(t, expected.ID, actual.ID) {
		return false
	}
	if !assert.Equal(t, expected.CalendarID, actual.CalendarID) {
		return false
	}
	if !assert.Equal(t, expected.Summary, actual.Summary) {
		return false
	}
	if !assertEqualTime(t, expected.Start, actual.Start) {
		return false
	}
	if !assertEqualTime(t, expected.End, actual.End) {
		return false
	}
	if !assert.Equal(t, expected.Status, actual.Status) {
		return false
	}

	return true
}
