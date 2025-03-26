package domain

import (
	"fmt"
	"net/http"
)

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
	RequiredError = func(paramName string) *ClientError {
		return NewClientError(http.StatusBadRequest, fmt.Sprintf("%s is required", paramName))
	}
	NotAllowedError = func(paramName string) *ClientError {
		return NewClientError(http.StatusBadRequest, fmt.Sprintf("%s is not allowed", paramName))
	}

	InvalidJSONError          = NewClientError(http.StatusBadRequest, "invalid json")
	CalendarNotFoundError     = NewClientError(http.StatusNotFound, "calender not found")
	CalendarAlreadyExistError = NewClientError(http.StatusNotFound, "calender already exists")
	AllParameterFalseError    = NewClientError(http.StatusBadRequest, "all must be true")
)
