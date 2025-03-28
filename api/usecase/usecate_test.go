package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/takuoki/golib/applog"
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

	exitCode := m.Run()

	// Truncate all data from existing tables
	tables := []string{
		"sync_histories",
		"channel_histories",
		"events",
		"calendars",
	}
	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			panic("fail to delete from " + table + ": " + err.Error())
		}
	}

	os.Exit(exitCode)
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
