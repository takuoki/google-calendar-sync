package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/openapi"
)

func (h *handler) PostCalendarsCalendarId(c echo.Context, calendarID string) error {
	ctx := context.Background()

	var req openapi.PostCalendarsCalendarIdJSONBody
	if err := c.Bind(&req); err != nil {
		// TODO: ClientError
		return fmt.Errorf("invalid request body: %w", err)
	}

	if req.Name == nil {
		// TODO: ClientError
		return fmt.Errorf("name is required")
	}

	err := h.calendarUsecase.Create(ctx, valueobject.CalendarID(calendarID), *req.Name, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("fail to create calendar: %w", err)
	}

	return success(c)
}
