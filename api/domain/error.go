package domain

import (
	"fmt"
	"net/http"
)

type ClientError struct {
	Code    int
	Message string
}

func newClientError(code int, message string) *ClientError {
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
		return newClientError(http.StatusBadRequest, fmt.Sprintf("%s is required", paramName))
	}
	NotAllowedError = func(paramName string) *ClientError {
		return newClientError(http.StatusBadRequest, fmt.Sprintf("%s is not allowed", paramName))
	}

	InvalidJSONError          = newClientError(http.StatusBadRequest, "invalid json")
	CalendarNotFoundError     = newClientError(http.StatusNotFound, "calender not found")
	CalendarAlreadyExistError = newClientError(http.StatusNotFound, "calender already exists")
	AllParameterFalseError    = newClientError(http.StatusBadRequest, "all must be true")
)

// InternalHandlingError is an error used for internal handling.
type InternalHandlingError struct {
	Message string
}

func newInternalHandlingError(message string) *InternalHandlingError {
	return &InternalHandlingError{
		Message: message,
	}
}

func (e *InternalHandlingError) Error() string {
	return e.Message
}

var (
	SyncTokenIsOldError = newInternalHandlingError("sync token is old")
)
