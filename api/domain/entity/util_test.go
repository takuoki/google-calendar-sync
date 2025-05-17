package entity_test

import (
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
)

func TestCompareTime(t *testing.T) {
	t.Parallel()

	now := time.Now()
	otherTime := now.Add(time.Hour)

	tests := map[string]struct {
		t1       *time.Time
		t2       *time.Time
		expected bool
	}{
		"both nil": {
			t1:       nil,
			t2:       nil,
			expected: true,
		},
		"one nil": {
			t1:       &now,
			t2:       nil,
			expected: false,
		},
		"equal times": {
			t1:       &now,
			t2:       &now,
			expected: true,
		},
		"different times": {
			t1:       &now,
			t2:       &otherTime,
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := entity.CompareTime(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("CompareTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}
