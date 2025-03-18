package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/openapi"
)

func (h *handler) PostWatch(c echo.Context, params openapi.PostWatchParams) error {
	ctx := context.Background()

	if params.All != nil && !*params.All {
		return domain.AllParameterFalseError
	}

	err := h.watchUsecase.StartAll(ctx)
	if err != nil {
		return fmt.Errorf("fail to watch all calendars: %w", err)
	}

	return success(c)
}

func (h *handler) PostWatchCalendarId(c echo.Context, calendarId string) error {
	ctx := context.Background()

	err := h.watchUsecase.Start(ctx, valueobject.CalendarID(calendarId))
	if err != nil {
		return fmt.Errorf("fail to watch calendar: %w", err)
	}

	return success(c)
}

func (h *handler) DeleteWatchCalendarId(c echo.Context, calendarId string) error {
	ctx := context.Background()

	err := h.watchUsecase.Stop(ctx, valueobject.CalendarID(calendarId))
	if err != nil {
		return fmt.Errorf("fail to stop watch: %w", err)
	}

	return success(c)
}
