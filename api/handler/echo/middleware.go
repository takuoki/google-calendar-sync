package echo

import (
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain"
)

func ErrorMiddleware(logger applog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			if err := next(c); err != nil {
				var e *domain.ClientError
				if errors.As(err, &e) {
					return failure(c, e.Code, e.Message)
				} else {
					logger.Errorf(ctx, "unexpected error: %v", err)
					return failure(c, http.StatusInternalServerError, "internal server error")
				}
			}

			return nil
		}
	}
}
