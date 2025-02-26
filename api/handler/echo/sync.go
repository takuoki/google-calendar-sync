package echo

import (
	"context"
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (h *handler) PostSyncCalendarId(c echo.Context, calendarID string) error {
	ctx := context.Background()

	// 更新検知時以外（Sync開始時等）は何もしない
	resourceState := c.Request().Header.Get("X-Goog-Resource-State")
	if resourceState != "exists" {
		h.logger.Infof(ctx, "skip (resource state: %q)", resourceState)
		return skipped(c)
	}

	if err := h.syncUsecase.Sync(ctx, valueobject.CalendarID(calendarID)); err != nil {
		return fmt.Errorf("fail to sync calendar: %w", err)
	}

	return success(c)
}
