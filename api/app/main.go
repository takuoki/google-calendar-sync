package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	calendar "google.golang.org/api/calendar/v3"

	applog "github.com/takuoki/golib/applog"
	echo_recovery "github.com/takuoki/golib/middleware/http/echo/recovery"
	echo_requestlog "github.com/takuoki/golib/middleware/http/echo/requestlog"
	"github.com/takuoki/golib/recovery"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	echohandler "github.com/takuoki/google-calendar-sync/api/handler/echo"
	"github.com/takuoki/google-calendar-sync/api/openapi"
	"github.com/takuoki/google-calendar-sync/api/repository/cloudsql"
	"github.com/takuoki/google-calendar-sync/api/repository/googlecalendar"
	"github.com/takuoki/google-calendar-sync/api/repository/mysql"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func main() {
	exitCode := 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	ctx := context.Background()

	logger := applog.NewBasicLogger(
		os.Stdout,
		applog.LevelOption(
			applog.ParseLevelWithDefault(os.Getenv("LOG_LEVEL"), applog.InfoLevel),
		),
	)

	// Database
	db, err := connectDB()
	if err != nil {
		logger.Criticalf(ctx, "fail to connect db: %v", err)
		exitCode = 1
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Criticalf(ctx, "fail to close db: %v", err)
		}
	}()

	if err := waitForDatabaseReady(ctx, db); err != nil {
		logger.Criticalf(ctx, "fail to wait for db ready: %v", err)
		exitCode = 2
		return
	}

	// Application
	calendarService, err := calendar.NewService(ctx)
	if err != nil {
		logger.Criticalf(ctx, "fail to create calendar service: %v", err)
		exitCode = 3
		return
	}
	clockService := service.NewClock()

	googleCalendarRepo := googlecalendar.NewGoogleCalendarRepository(
		os.Getenv("WEBHOOK_BASE_URL"), calendarService, clockService, logger)
	mysqlRepo := mysql.NewMysqlRepository(db, clockService, logger)

	calendarUsecase := usecase.NewCalendarUsecase(mysqlRepo, logger)
	syncUsecase := usecase.NewSyncUsecase(clockService, googleCalendarRepo, mysqlRepo, logger)
	watchUsecase := usecase.NewWatchUsecase(googleCalendarRepo, mysqlRepo, logger)

	handler := echohandler.New(calendarUsecase, syncUsecase, watchUsecase, logger)

	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.AddTrailingSlash())

	e.Use(echo_requestlog.Middleware(logger))
	e.Use(echohandler.ErrorMiddleware(logger))
	e.Use(echo_recovery.Middleware(echo_recovery.RecoveryFunc(recovery.Recovery)))

	eg := e.Group("/api")
	openapi.RegisterHandlers(eg, handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof(ctx, "listening and serving on port %s", port)

	if err := e.Start(":" + port); err != nil {
		logger.Criticalf(ctx, "failed to serve: %v", err)
		exitCode = 4
		return
	}
}

func connectDB() (*sql.DB, error) {
	switch os.Getenv("DB_TYPE") {
	case "cloudsql":
		// TODO: 環境変数はここで展開する
		return cloudsql.ConnectWithConnector()
	case "mysql":
		return mysql.ConnectDB(
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
	default:
		return nil, fmt.Errorf("unknown db type: %q", os.Getenv("DB_TYPE"))
	}
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
