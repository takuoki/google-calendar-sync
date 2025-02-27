package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (h *handler) PostSyncCalendarId(c echo.Context, calendarID string) error {
	ctx := context.Background()

	if err := h.syncUsecase.Sync(ctx, valueobject.CalendarID(calendarID)); err != nil {
		return fmt.Errorf("fail to sync calendar: %w", err)
	}

	return success(c)
}
