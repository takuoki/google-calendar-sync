package service_test

import (
	"strings"
	"testing"

	"github.com/takuoki/google-calendar-sync/api/domain/service"
)

func TestNewSystemClock_InvalidLocation(t *testing.T) {
	_, err := service.NewSystemClock("Invalid/Location")
	if err == nil {
		t.Error("expected error for invalid location, got nil")
	} else if !strings.HasPrefix(err.Error(), "fail to load location") {
		t.Errorf("error message does not match the expected format, got: %s", err.Error())
	}
}

func TestSystemClock_Now(t *testing.T) {
	location := "UTC"
	clock, err := service.NewSystemClock(location)
	if err != nil {
		t.Fatalf("failed to create SystemClock: %v", err)
	}

	now := clock.Now()
	if now.IsZero() {
		t.Error("expected non-zero time, got zero value")
	}
}

func TestSystemClock_Today(t *testing.T) {
	location := "UTC"
	clock, err := service.NewSystemClock(location)
	if err != nil {
		t.Fatalf("failed to create SystemClock: %v", err)
	}

	today := clock.Today()
	if today.IsZero() {
		t.Error("expected non-zero time, got zero value")
	}
}
