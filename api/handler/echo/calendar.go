package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	"github.com/takuoki/google-calendar-sync/api/openapi"
)

func (h *handler) PostCalendarsCalendarId(c echo.Context, calendarID string) error {
	ctx := context.Background()

	var req openapi.PostCalendarsCalendarIdJSONBody
	if err := c.Bind(&req); err != nil {
		return domain.InvalidJSONError
	}

	if req.Name == nil {
		return domain.RequiredError("name")
	}

	err := h.calendarUsecase.Create(ctx, valueobject.CalendarID(calendarID), *req.Name, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("fail to create calendar: %w", err)
	}

	return success(c)
}
