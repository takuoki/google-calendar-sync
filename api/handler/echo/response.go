package echo

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func success(c echo.Context) error {
	response := Response{
		Status: "success",
	}
	return c.JSON(http.StatusOK, response)
}

func failure(c echo.Context, status int, message string) error {
	response := Response{
		Status:  "error",
		Message: message,
	}
	return c.JSON(status, response)
}
