package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/openapi"
)

func (h *handler) PostSyncCalendarId(c echo.Context, calendarID string) error {
	ctx := context.Background()

	if err := h.syncUsecase.Sync(ctx, valueobject.CalendarID(calendarID)); err != nil {
		return fmt.Errorf("fail to sync calendar: %w", err)
	}

	return success(c)
}

func (h *handler) PostSyncFutureInstance(c echo.Context, params openapi.PostSyncFutureInstanceParams) error {
	ctx := context.Background()

	if params.All != nil && !*params.All {
		return domain.AllParameterFalseError
	}

	if err := h.syncUsecase.SyncFutureInstanceAll(ctx); err != nil {
		return fmt.Errorf("fail to sync future instance: %w", err)
	}

	return success(c)
}
