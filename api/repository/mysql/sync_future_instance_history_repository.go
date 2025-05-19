package mysql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (tx *mysqlTransaction) CreateSyncFutureInstanceHistory(
	ctx context.Context, calendarID valueobject.CalendarID, syncTime time.Time,
	updatedEventCount int) error {

	err := createSyncFutureInstanceHistory(ctx, tx.tx, calendarID, syncTime, updatedEventCount)
	if err != nil {
		return fmt.Errorf("fail to create sync future instance history: %w", err)
	}

	return nil
}

func createSyncFutureInstanceHistory(
	ctx context.Context, db database, calendarID valueobject.CalendarID, syncTime time.Time,
	updatedEventCount int) error {

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO sync_future_instance_histories "+
			"(calendar_id, sync_time, updated_event_count) "+
			"VALUES (?, ?, ?)",
		calendarID, syncTime, updatedEventCount)

	if err != nil {
		return fmt.Errorf("fail to insert sync future instance history: %w", err)
	}

	return nil
}

func (r *MysqlRepository) DeleteAllSyncFutureInstanceHistoriesForMain(ctx context.Context, m *testing.M) (updatedCount int, err error) {
	updatedCount, err = r.deleteAllSyncFutureInstanceHistories(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync future instance histories: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) DeleteAllSyncFutureInstanceHistories(ctx context.Context, t *testing.T) (updatedCount int, err error) {
	t.Helper()

	updatedCount, err = r.deleteAllSyncFutureInstanceHistories(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync future instance histories: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) deleteAllSyncFutureInstanceHistories(ctx context.Context) (updatedCount int, err error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sync_future_instance_histories")
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync future instance histories: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to get affected rows: %w", err)
	}
	updatedCount = int(affectedRows)

	return updatedCount, nil
}
