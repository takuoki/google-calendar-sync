package echo

import (
	"github.com/takuoki/golib/applog"

	"github.com/takuoki/google-calendar-sync/api/openapi"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

type handler struct {
	calendarUsecase usecase.CalendarUsecase
	syncUsecase     usecase.SyncUsecase
	watchUsecase    usecase.WatchUsecase
	logger          applog.Logger
}

func New(
	calendarUsecase usecase.CalendarUsecase,
	syncUsecase usecase.SyncUsecase,
	watchUsecase usecase.WatchUsecase,
	logger applog.Logger,
) openapi.ServerInterface {
	return &handler{
		calendarUsecase: calendarUsecase,
		syncUsecase:     syncUsecase,
		watchUsecase:    watchUsecase,
		logger:          logger,
	}
}
