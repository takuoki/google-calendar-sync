package mysql

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/constant"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *MysqlRepository) ListEvents(ctx context.Context, t *testing.T,
	calendarID valueobject.CalendarID) ([]entity.Event, error) {
	t.Helper()

	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, calendar_id, summary, start, end, status "+
			"FROM events "+
			"WHERE calendar_id = ? "+
			"ORDER BY id",
		calendarID)
	if err != nil {
		return nil, fmt.Errorf("fail to select events: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Errorf(ctx, "fail to close rows: %s", closeErr)
		}
	}()

	var events []entity.Event
	for rows.Next() {
		var event entity.Event
		err := rows.Scan(&event.ID, &event.CalendarID, &event.Summary,
			&event.Start, &event.End, &event.Status)
		if err != nil {
			return nil, fmt.Errorf("fail to scan row: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *MysqlRepository) CreateEvent(ctx context.Context, t *testing.T,
	event entity.Event) error {
	t.Helper()

	err := createEvent(ctx, r.db, event)
	if err != nil {
		return fmt.Errorf("fail to create event: %w", err)
	}

	return nil
}

func createEvent(ctx context.Context, db database, event entity.Event) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO events "+
			"(calendar_id, id, recurring_event_id, summary, start, end, status) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		event.CalendarID, event.ID, event.RecurringEventID, event.Summary, event.Start, event.End, event.Status)
	if err != nil {
		return fmt.Errorf("fail to insert event: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) SyncEvents(ctx context.Context, calendarID valueobject.CalendarID, events []entity.Event) (int, error) {

	if len(events) == 0 {
		return 0, nil
	}

	eventIDs := make([]valueobject.EventID, 0, len(events))
	for _, event := range events {
		if event.CalendarID != calendarID {
			return 0, fmt.Errorf("all events must have the same calendar ID: expected %s, got %s", calendarID, event.CalendarID)
		}
		eventIDs = append(eventIDs, event.ID)
	}

	dbEventMap, err := tx.fetchEventMap(ctx, calendarID, eventIDs)
	if err != nil {
		return 0, err
	}

	// 更新順に影響するため ID でソート
	sort.Slice(events, func(i, j int) bool {
		return events[i].ID < events[j].ID
	})

	// TODO: バッチ処理ができるとよい
	updateCount := 0
	for _, event := range events {
		dbEvent, ok := dbEventMap[event.ID]

		// DB に存在しない場合は挿入
		if !ok {
			if err := createEvent(ctx, tx.tx, event); err != nil {
				return 0, fmt.Errorf("fail to create event: %w", err)
			}
			updateCount++
			continue
		}

		// DB に存在するが、データが同じ場合はスキップ
		if event.Equals(&dbEvent) {
			continue
		}

		// DB に存在するが、内容が異なる場合は更新
		cnt, err := tx.updateEvent(ctx, event)
		if err != nil {
			return 0, fmt.Errorf("fail to update event: %w", err)
		}
		updateCount += cnt
	}

	return updateCount, nil
}

func (tx *mysqlTransaction) fetchEventMap(ctx context.Context,
	calendarID valueobject.CalendarID, eventIDs []valueobject.EventID) (map[valueobject.EventID]entity.Event, error) {

	if len(eventIDs) == 0 {
		return nil, nil
	}

	tmpEventIDs := make([]any, 0, len(eventIDs))
	placeholders := make([]string, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		tmpEventIDs = append(tmpEventIDs, eventID)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"SELECT id, calendar_id, recurring_event_id, summary, start, end, status "+
			"FROM events WHERE calendar_id = ? AND id IN (%s)",
		strings.Join(placeholders, ", "),
	)

	rows, err := tx.tx.QueryContext(ctx, query, append([]any{calendarID}, tmpEventIDs...)...)
	if err != nil {
		return nil, fmt.Errorf("fail to select events: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			tx.logger.Errorf(ctx, "fail to close rows: %s", closeErr)
		}
	}()

	eventMap := map[valueobject.EventID]entity.Event{}
	for rows.Next() {
		var event entity.Event
		err := rows.Scan(
			&event.ID,
			&event.CalendarID,
			&event.RecurringEventID,
			&event.Summary,
			&event.Start,
			&event.End,
			&event.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("fail to scan event: %w", err)
		}
		eventMap[event.ID] = event
	}

	return eventMap, nil
}

func (tx *mysqlTransaction) updateEvent(ctx context.Context, event entity.Event) (updatedCount int, err error) {
	result, err := tx.tx.ExecContext(ctx,
		"UPDATE events SET recurring_event_id = ?, summary = ?, start = ?, end = ?, status = ? "+
			"WHERE calendar_id = ? AND id = ?",
		event.RecurringEventID, event.Summary, event.Start, event.End, event.Status, event.CalendarID, event.ID)
	if err != nil {
		return 0, fmt.Errorf("fail to update event: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to get affected rows: %w", err)
	}
	updatedCount = int(affectedRows)

	return updatedCount, nil
}

func (tx *mysqlTransaction) cancelEventInstancesWithAfter(ctx context.Context,
	calendarID valueobject.CalendarID, recurringEventID valueobject.EventID, excludedEventIDs []valueobject.EventID,
	after time.Time) (updatedCount int, err error) {

	var query string
	var args []interface{}

	if len(excludedEventIDs) == 0 {
		// excludedEventIDsが空の場合はid NOT IN 句は不要
		query = "UPDATE events SET status = ? " +
			"WHERE calendar_id = ? AND recurring_event_id = ? AND start >= ?"
		args = []interface{}{constant.EventStatusCancelled, calendarID, recurringEventID, after}
	} else {
		placeholders := make([]string, len(excludedEventIDs))
		for i := range excludedEventIDs {
			placeholders[i] = "?"
		}
		query = "UPDATE events SET status = ? " +
			"WHERE calendar_id = ? AND recurring_event_id = ? AND id NOT IN (" +
			strings.Join(placeholders, ",") + ") AND start >= ?"
		args = append(args, constant.EventStatusCancelled, calendarID, recurringEventID)
		for _, id := range excludedEventIDs {
			args = append(args, id)
		}
		args = append(args, after)
	}

	result, err := tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("fail to update events: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to get affected rows: %w", err)
	}
	updatedCount = int(affectedRows)

	return updatedCount, nil
}

func (r *MysqlRepository) DeleteAllEventsForMain(ctx context.Context, m *testing.M) (updatedCount int, err error) {
	updatedCount, err = r.deleteAllEvents(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all events: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) DeleteAllEvents(ctx context.Context, t *testing.T) (updatedCount int, err error) {
	t.Helper()

	updatedCount, err = r.deleteAllEvents(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all events: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) deleteAllEvents(ctx context.Context) (updatedCount int, err error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM events")
	if err != nil {
		return 0, fmt.Errorf("fail to delete all events: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to get affected rows: %w", err)
	}
	updatedCount = int(affectedRows)

	return updatedCount, nil
}
