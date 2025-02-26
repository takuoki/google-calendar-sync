package echo

import (
	"github.com/takuoki/golib/applog"

	"github.com/takuoki/google-calendar-sync/api/openapi"
	"github.com/takuoki/google-calendar-sync/api/usecase"
)

type handler struct {
	syncUsecase  usecase.SyncUsecase
	watchUsecase usecase.WatchUsecase
	logger       applog.Logger
}

func New(
	syncUsecase usecase.SyncUsecase,
	watchUsecase usecase.WatchUsecase,
	logger applog.Logger,
) openapi.ServerInterface {
	return &handler{
		syncUsecase:  syncUsecase,
		watchUsecase: watchUsecase,
		logger:       logger,
	}
}
