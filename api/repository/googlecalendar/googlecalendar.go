package googlecalendar

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

type googleCalendarRepository struct {
	webhookBaseURL string
	service        *calendar.Service
	clockService   service.Clock
	logger         applog.Logger
}

func NewGoogleCalendarRepository(ctx context.Context, webhookBaseURL string,
	clockService service.Clock, logger applog.Logger) (repository.GoogleCalendarRepository, error) {

	if webhookBaseURL == "" {
		return nil, fmt.Errorf("webhook base url is required")
	}

	service, err := calendar.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to create calendar service: %w", err)
	}

	return &googleCalendarRepository{
		webhookBaseURL: webhookBaseURL,
		service:        service,
		clockService:   clockService,
		logger:         logger,
	}, nil
}

type RefreshTokenResolver interface {
	GetRefreshToken(ctx context.Context, calendarID valueobject.CalendarID) (string, error)
}

type googleCalendarWithOauthRepository struct {
	webhookBaseURL       string
	oauth2Config         *oauth2.Config
	refreshTokenResolver RefreshTokenResolver
	clockService         service.Clock
	logger               applog.Logger
}

func NewGoogleCalendarWithOauthRepository(webhookBaseURL, oauthClientID, oauthClientSecret, oauthRedirectURL string,
	refreshTokenResolver RefreshTokenResolver, clockService service.Clock, logger applog.Logger) (repository.GoogleCalendarRepository, error) {

	if webhookBaseURL == "" {
		return nil, fmt.Errorf("webhook base url is required")
	}

	if oauthClientID == "" {
		return nil, fmt.Errorf("oauth client id is required")
	}

	if oauthClientSecret == "" {
		return nil, fmt.Errorf("oauth client secret is required")
	}

	if oauthRedirectURL == "" {
		return nil, fmt.Errorf("oauth redirect url is required")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  oauthRedirectURL,
	}

	return &googleCalendarWithOauthRepository{
		webhookBaseURL:       webhookBaseURL,
		oauth2Config:         oauth2Config,
		refreshTokenResolver: refreshTokenResolver,
		clockService:         clockService,
		logger:               logger,
	}, nil
}

func (r *googleCalendarWithOauthRepository) getCalendarService(ctx context.Context, calendarID valueobject.CalendarID) (*calendar.Service, error) {

	refreshToken, err := r.refreshTokenResolver.GetRefreshToken(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("fail to get refresh token: %w", err)
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := r.oauth2Config.TokenSource(ctx, token)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("fail to create calendar service: %w", err)
	}

	return calendarService, nil
}

func convertDateTime(datetime *calendar.EventDateTime, location string) (*time.Time, error) {
	if datetime == nil {
		return nil, nil
	}

	// 終日指定の場合は Date に設定されるため 00:00:00 の日時に変換する
	// タイムゾーンはカレンダーのタイムゾーンを使用する
	if datetime.Date != "" {

		loc, err := time.LoadLocation(location)
		if err != nil {
			return nil, fmt.Errorf("fail to load location: %w", err)
		}

		t, err := time.ParseInLocation("2006-01-02", datetime.Date, loc)
		if err != nil {
			return nil, fmt.Errorf("fail to parse date: %w", err)
		}

		return &t, nil
	}

	if datetime.DateTime != "" {
		t, err := time.Parse(time.RFC3339, datetime.DateTime)
		if err != nil {
			return nil, fmt.Errorf("fail to parse datetime: %w", err)
		}
		return &t, nil
	}

	return nil, fmt.Errorf("invalid datetime: %+v", datetime)
}

func convertUnitTime(t int64) (time.Time, error) {
	sec := t / 1000
	nsec := (t % 1000) * 1000000
	return time.Unix(sec, nsec), nil
}
