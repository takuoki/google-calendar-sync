package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/constant"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *MysqlRepository) ListActiveRecurringEvents(ctx context.Context, calendarID valueobject.CalendarID) (
	[]entity.RecurringEvent, error) {

	// TODO: recurrence の UNTIL を考慮し、有効期限が切れたイベントも除外したい
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, calendar_id, summary, recurrence, start, end, status "+
			"FROM recurring_events "+
			"WHERE calendar_id = ? AND status != ? "+
			"ORDER BY id",
		calendarID, constant.EventStatusCancelled)
	if err != nil {
		return nil, fmt.Errorf("fail to select recurring events: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Errorf(ctx, "fail to close rows: %s", closeErr)
		}
	}()

	var recurringEvents []entity.RecurringEvent
	for rows.Next() {
		var recurringEvent entity.RecurringEvent
		err := rows.Scan(&recurringEvent.ID, &recurringEvent.CalendarID, &recurringEvent.Summary,
			&recurringEvent.Recurrence, &recurringEvent.Start, &recurringEvent.End, &recurringEvent.Status)
		if err != nil {
			return nil, fmt.Errorf("fail to scan row: %w", err)
		}

		recurringEvents = append(recurringEvents, recurringEvent)
	}

	return recurringEvents, nil
}

func (tx *mysqlTransaction) SyncRecurringEventAndInstancesWithAfter(ctx context.Context,
	recurringEvent entity.RecurringEvent, instances []entity.Event, after time.Time) (updatedCount int, err error) {

	updatedCount, err = tx.syncRecurringEvent(ctx, recurringEvent)
	if err != nil {
		return 0, fmt.Errorf("fail to sync recurring event: %w", err)
	}

	cnt, err := tx.SyncEvents(ctx, recurringEvent.CalendarID, instances)
	if err != nil {
		return 0, fmt.Errorf("fail to sync events: %w", err)
	}

	updatedCount += cnt

	// DB に存在するが、引数で渡されなかったイベントは削除
	eventIDs := make([]valueobject.EventID, 0, len(instances))
	for _, event := range instances {
		eventIDs = append(eventIDs, event.ID)
	}

	cnt, err = tx.cancelRecurringEventInstancesWithAfter(
		ctx, recurringEvent.CalendarID, recurringEvent.ID, eventIDs, after)
	if err != nil {
		return 0, fmt.Errorf("fail to cancel recurring event instances: %w", err)
	}

	updatedCount += cnt

	return updatedCount, nil
}

func (tx *mysqlTransaction) syncRecurringEvent(ctx context.Context, recurringEvent entity.RecurringEvent) (
	updatedCount int, err error) {

	var dbRecurringEvent entity.RecurringEvent

	err = tx.tx.QueryRowContext(
		ctx,
		"SELECT id, calendar_id, summary, recurrence, start, end, status "+
			"FROM recurring_events WHERE id = ?",
		recurringEvent.ID,
	).Scan(&dbRecurringEvent.ID, &dbRecurringEvent.CalendarID, &dbRecurringEvent.Summary,
		&dbRecurringEvent.Recurrence, &dbRecurringEvent.Start, &dbRecurringEvent.End, &dbRecurringEvent.Status)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("fail to select recurring event: %w", err)
	}

	// DB に存在しない場合は挿入
	if err == sql.ErrNoRows {
		if err := tx.createRecurringEvent(ctx, recurringEvent); err != nil {
			return 0, fmt.Errorf("fail to create recurring event: %w", err)
		}
		return 1, nil
	}

	// DB に存在するが、データが同じ場合はスキップ
	if recurringEvent.Equals(&dbRecurringEvent) {
		return 0, nil
	}

	// DB に存在するが、データが異なる場合は更新
	if err := tx.updateRecurringEvent(ctx, recurringEvent); err != nil {
		return 0, fmt.Errorf("fail to update recurring event: %w", err)
	}

	return 1, nil
}

func (tx *mysqlTransaction) createRecurringEvent(ctx context.Context, recurringEvent entity.RecurringEvent) error {
	_, err := tx.tx.ExecContext(
		ctx,
		"INSERT INTO recurring_events "+
			"(id, calendar_id, summary, recurrence, start, end, status) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		recurringEvent.ID, recurringEvent.CalendarID, recurringEvent.Summary,
		recurringEvent.Recurrence, recurringEvent.Start, recurringEvent.End, recurringEvent.Status)
	if err != nil {
		return fmt.Errorf("fail to insert recurring event: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) updateRecurringEvent(ctx context.Context, recurringEvent entity.RecurringEvent) error {
	_, err := tx.tx.ExecContext(
		ctx,
		"UPDATE recurring_events "+
			"SET summary = ?, recurrence = ?, start = ?, end = ?, status = ? "+
			"WHERE id = ? AND calendar_id = ?",
		recurringEvent.Summary, recurringEvent.Recurrence, recurringEvent.Start,
		recurringEvent.End, recurringEvent.Status, recurringEvent.ID, recurringEvent.CalendarID)
	if err != nil {
		return fmt.Errorf("fail to update recurring event: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) cancelRecurringEventInstancesWithAfter(ctx context.Context,
	calendarID valueobject.CalendarID, recurringEventID valueobject.EventID, excludedEventIDs []valueobject.EventID,
	after time.Time) (updatedCount int, err error) {
	result, err := tx.tx.ExecContext(
		ctx,
		"UPDATE events SET status = ? "+
			"WHERE calendar_id = ? AND recurring_event_id = ? AND id NOT IN (?) AND start >= ?",
		constant.EventStatusCancelled, calendarID, recurringEventID, excludedEventIDs, after)
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
