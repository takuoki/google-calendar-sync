package mysql

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (tx *mysqlTransaction) SyncEvents(ctx context.Context, events []entity.Event) (int, error) {

	if len(events) == 0 {
		return 0, nil
	}

	eventIDs := make([]valueobject.EventID, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID)
	}

	dbEventMap, err := tx.fetchEventMapWithLock(ctx, eventIDs)
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
			if err := tx.insertEvent(ctx, event); err != nil {
				return 0, fmt.Errorf("fail to insert event: %w", err)
			}
			updateCount++
			continue
		}

		// DB に存在するが、データが同じ場合はスキップ
		if event.Equals(&dbEvent) {
			continue
		}

		// DB に存在するが、内容が異なる場合は更新
		if err := tx.updateEvent(ctx, event); err != nil {
			return 0, fmt.Errorf("fail to update event: %w", err)
		}
		updateCount++
	}

	return updateCount, nil
}

func (tx *mysqlTransaction) fetchEventMapWithLock(
	ctx context.Context, eventIDs []valueobject.EventID) (map[valueobject.EventID]entity.Event, error) {

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
		"SELECT id, calendar_id, summary, start, end, status "+
			"FROM events WHERE id IN (%s) "+
			"FOR UPDATE",
		strings.Join(placeholders, ", "),
	)

	rows, err := tx.tx.QueryContext(ctx, query, tmpEventIDs...)
	if err != nil {
		return nil, fmt.Errorf("fail to select events: %w", err)
	}
	defer rows.Close()

	eventMap := map[valueobject.EventID]entity.Event{}
	for rows.Next() {
		var event entity.Event
		err := rows.Scan(
			&event.ID,
			&event.CalendarID,
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

func (tx *mysqlTransaction) insertEvent(ctx context.Context, event entity.Event) error {
	_, err := tx.tx.ExecContext(ctx,
		"INSERT INTO events "+
			"(id, calendar_id, summary, start, end, status) "+
			"VALUES (?, ?, ?, ?, ?, ?)",
		event.ID, event.CalendarID, event.Summary, event.Start, event.End, event.Status)
	if err != nil {
		return fmt.Errorf("fail to insert event: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) updateEvent(ctx context.Context, event entity.Event) error {
	_, err := tx.tx.ExecContext(ctx,
		// CalendarID は更新しない
		"UPDATE events SET summary = ?, start = ?, end = ?, status = ? WHERE id = ?",
		event.Summary, event.Start, event.End, event.Status, event.ID)
	if err != nil {
		return fmt.Errorf("fail to update event: %w", err)
	}

	return nil
}
