package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/openapi"
)

func (h *handler) PostCalendarsCalendarId(
	c echo.Context, calendarID string, params openapi.PostCalendarsCalendarIdParams) error {
	ctx := context.Background()

	if err := h.calendarUsecase.Create(ctx, valueobject.CalendarID(calendarID), params.Name); err != nil {
		return fmt.Errorf("fail to sync calendar: %w", err)
	}

	return success(c)
}
