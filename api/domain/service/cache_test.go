package service_test

import (
	"testing"

	"github.com/takuoki/google-calendar-sync/api/domain/service"
)

func TestInMemoryCache(t *testing.T) {
	cache := service.NewInMemoryCache[string, int]()

	// Test Set and Get
	cache.Set("key1", 100)
	if value, ok := cache.Get("key1"); !ok || value != 100 {
		t.Errorf("expected value 100, got %v (exists: %v)", value, ok)
	}

	// Test Get for non-existent key
	if _, ok := cache.Get("key2"); ok {
		t.Errorf("expected key2 to not exist")
	}

	// Test Delete
	cache.Delete("key1")
	if _, ok := cache.Get("key1"); ok {
		t.Errorf("expected key1 to be deleted")
	}
}
