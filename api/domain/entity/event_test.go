package entity_test

import (
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func TestEvent_Equals(t *testing.T) {
	t.Parallel()

	now := time.Now()
	otherTime := now.Add(time.Hour)

	tests := map[string]struct {
		event1   *entity.Event
		event2   *entity.Event
		expected bool
	}{
		"both nil": {
			event1:   nil,
			event2:   nil,
			expected: true,
		},
		"one nil": {
			event1:   &entity.Event{},
			event2:   nil,
			expected: false,
		},
		"equal events": {
			event1: &entity.Event{
				ID:               valueobject.EventID("1"),
				CalendarID:       valueobject.CalendarID("cal1"),
				RecurringEventID: valueobject.NewEventID("r1"),
				Summary:          "Meeting",
				Start:            &now,
				End:              &otherTime,
				Status:           "confirmed",
			},
			event2: &entity.Event{
				ID:               valueobject.EventID("1"),
				CalendarID:       valueobject.CalendarID("cal1"),
				RecurringEventID: valueobject.NewEventID("r1"),
				Summary:          "Meeting",
				Start:            &now,
				End:              &otherTime,
				Status:           "confirmed",
			},
			expected: true,
		},
		"different ID": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("2"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different CalendarID": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal2"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different RecurringEventID": {
			event1: &entity.Event{
				ID:               valueobject.EventID("1"),
				CalendarID:       valueobject.CalendarID("cal1"),
				RecurringEventID: valueobject.NewEventID("r1"),
				Summary:          "Meeting",
				Start:            &now,
				End:              &otherTime,
				Status:           "confirmed",
			},
			event2: &entity.Event{
				ID:               valueobject.EventID("1"),
				CalendarID:       valueobject.CalendarID("cal2"),
				RecurringEventID: valueobject.NewEventID("r2"),
				Summary:          "Meeting",
				Start:            &now,
				End:              &otherTime,
				Status:           "confirmed",
			},
			expected: false,
		},
		"different Summary": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Workshop",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Start": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &otherTime,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different End": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &now,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Status": {
			event1: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.Event{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Start:      &now,
				End:        &otherTime,
				Status:     "cancelled",
			},
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tt.event1.Equals(tt.event2)
			if result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}
