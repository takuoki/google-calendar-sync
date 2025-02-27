package main

import (
	"context"
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
	"github.com/takuoki/google-calendar-sync/api/repository/googlecalendar"
	"github.com/takuoki/google-calendar-sync/api/repository/mysql"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

const (
	pingRetryLimit    = 30
	pingRetryInterval = 1 * time.Second
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
	db, err := mysql.ConnectDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
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

	for i := 0; ; i++ {
		if i >= pingRetryLimit {
			logger.Criticalf(ctx, "fail to ping db: %v", err)
			exitCode = 2
			return
		}
		if err := db.Ping(); err != nil {
			break
		}
		time.Sleep(pingRetryInterval)
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
