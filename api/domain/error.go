package domain

import "net/http"

type ClientError struct {
	Code    int
	Message string
}

func NewClientError(code int, message string) *ClientError {
	return &ClientError{
		Code:    code,
		Message: message,
	}
}

func (e *ClientError) Error() string {
	return e.Message
}

var (
	CalendarNotFoundError     = NewClientError(http.StatusNotFound, "calender not found")
	CalendarAlreadyExistError = NewClientError(http.StatusNotFound, "calender already exists")
	AllParameterFalseError    = NewClientError(http.StatusBadRequest, "all must be true")
)
