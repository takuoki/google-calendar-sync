package entity_test

import (
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func TestRecurringEvent_Equals(t *testing.T) {
	t.Parallel()

	now := time.Now()
	otherTime := now.Add(time.Hour)

	tests := map[string]struct {
		event1   *entity.RecurringEvent
		event2   *entity.RecurringEvent
		expected bool
	}{
		"both nil": {
			event1:   nil,
			event2:   nil,
			expected: true,
		},
		"one nil": {
			event1:   &entity.RecurringEvent{},
			event2:   nil,
			expected: false,
		},
		"equal events": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: true,
		},
		"different ID": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("2"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different CalendarID": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal2"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different RecurringEventID": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal2"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Summary": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Workshop",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Recurrence": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Workshop",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO:UNTIL=20301231"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Start": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &otherTime,
				End:        &otherTime,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different End": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &now,
				Status:     "confirmed",
			},
			expected: false,
		},
		"different Status": {
			event1: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
				Start:      &now,
				End:        &otherTime,
				Status:     "confirmed",
			},
			event2: &entity.RecurringEvent{
				ID:         valueobject.EventID("1"),
				CalendarID: valueobject.CalendarID("cal1"),
				Summary:    "Meeting",
				Recurrence: `["RRULE:FREQ=WEEKLY;BYDAY=MO"]`,
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
