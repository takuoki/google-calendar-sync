package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	applog "github.com/takuoki/golib/applog"
	echo_recovery "github.com/takuoki/golib/middleware/http/echo/recovery"
	echo_requestlog "github.com/takuoki/golib/middleware/http/echo/requestlog"
	"github.com/takuoki/golib/recovery"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	echohandler "github.com/takuoki/google-calendar-sync/api/handler/echo"
	"github.com/takuoki/google-calendar-sync/api/openapi"
	"github.com/takuoki/google-calendar-sync/api/repository"
	"github.com/takuoki/google-calendar-sync/api/repository/cloudsql"
	"github.com/takuoki/google-calendar-sync/api/repository/googlecalendar"
	"github.com/takuoki/google-calendar-sync/api/repository/mysql"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

func main() {
	ctx := context.Background()

	logger := applog.NewBasicLogger(
		os.Stdout,
		applog.LevelOption(
			applog.ParseLevelWithDefault(os.Getenv("LOG_LEVEL"), applog.InfoLevel),
		),
	)

	if code, err := run(ctx, logger); err != nil {
		logger.Criticalf(ctx, "fail to run: %v", err)
		os.Exit(code)
	}
}

func run(ctx context.Context, logger applog.Logger) (exitCode int, er error) {

	db, err := connectDB()
	if err != nil {
		return 1, fmt.Errorf("fail to connect db: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Criticalf(ctx, "fail to close db: %v", err)
		}
	}()

	if err := waitForDatabaseReady(ctx, db); err != nil {
		return 2, fmt.Errorf("fail to wait for db ready: %w", err)
	}

	handler, err := setupApplication(ctx, db, logger)
	if err != nil {
		return 3, fmt.Errorf("fail to setup application: %w", err)
	}

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
		return 4, fmt.Errorf("fail to serve: %w", err)
	}

	return 0, nil
}

func connectDB() (*sql.DB, error) {
	switch os.Getenv("DB_TYPE") {
	case "cloudsql":
		return cloudsql.ConnectWithConnector(
			os.Getenv("INSTANCE_CONNECTION_NAME"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_USE_PRIVATE_IP"),
		)
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

func setupApplication(ctx context.Context, db *sql.DB, logger applog.Logger) (openapi.ServerInterface, error) {

	var err error

	oauthClientID := os.Getenv("OAUTH_CLIENT_ID")

	useOauth := false
	if oauthClientID != "" {
		useOauth = true
	}

	// Service
	clockService := service.NewClock()

	var cryptService *service.Crypt
	if useOauth {
		cryptService, err = service.NewCrypt([]byte(os.Getenv("CRYPT_KEY")))
		if err != nil {
			return nil, fmt.Errorf("fail to create crypt service: %w", err)
		}
	}

	// Repository
	mysqlRepo := mysql.NewMysqlRepository(db, clockService, cryptService, logger)

	var googleCalendarRepo repository.GoogleCalendarRepository
	if oauthClientID := os.Getenv("OAUTH_CLIENT_ID"); oauthClientID == "" {
		googleCalendarRepo, err = googlecalendar.NewGoogleCalendarRepository(
			ctx, os.Getenv("WEBHOOK_BASE_URL"), clockService, logger)
		if err != nil {
			return nil, fmt.Errorf("fail to create google calendar repository: %w", err)
		}
	} else {
		googleCalendarRepo, err = googlecalendar.NewGoogleCalendarWithOauthRepository(
			os.Getenv("WEBHOOK_BASE_URL"), oauthClientID, os.Getenv("OAUTH_CLIENT_SECRET"),
			os.Getenv("OAUTH_REDIRECT_URL"), mysqlRepo, clockService, logger)
		if err != nil {
			return nil, fmt.Errorf("fail to create google calendar with oauth repository: %w", err)
		}
	}

	// Usecase
	calendarUsecase := usecase.NewCalendarUsecase(mysqlRepo, useOauth, logger)
	syncUsecase := usecase.NewSyncUsecase(clockService, googleCalendarRepo, mysqlRepo, logger)
	watchUsecase := usecase.NewWatchUsecase(googleCalendarRepo, mysqlRepo, logger)

	// Handler
	handler := echohandler.New(calendarUsecase, syncUsecase, watchUsecase, logger)

	return handler, nil
}
