package googlecalendar

import (
	"fmt"
	"time"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/repository"
)

type googleCalendarRepository struct {
	webhookBaseURL string
	service        *calendar.Service
	clockService   service.Clock
	logger         applog.Logger
}

func NewGoogleCalendarRepository(webhookBaseURL string, service *calendar.Service,
	clockService service.Clock, logger applog.Logger) repository.GoogleCalendarRepository {
	return &googleCalendarRepository{
		webhookBaseURL: webhookBaseURL,
		service:        service,
		clockService:   clockService,
		logger:         logger,
	}
}

func convertDateTime(datetime *calendar.EventDateTime) (*time.Time, error) {
	if datetime == nil || datetime.DateTime == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, datetime.DateTime)
	if err != nil {
		return nil, fmt.Errorf("fail to parse datetime: %w", err)
	}
	return &t, nil
}

func convertUnitTime(t int64) (time.Time, error) {
	sec := t / 1000
	nsec := (t % 1000) * 1000000
	return time.Unix(sec, nsec), nil
}
