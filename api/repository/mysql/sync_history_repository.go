package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *MysqlRepository) GetLatestSyncToken(
	ctx context.Context, calendarID valueobject.CalendarID) (string, error) {

	var syncToken string

	err := r.db.QueryRowContext(
		ctx,
		"SELECT next_sync_token FROM sync_histories "+
			"WHERE calendar_id = ? "+
			"ORDER BY sync_time DESC LIMIT 1",
		calendarID,
	).Scan(&syncToken)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("fail to select latest sync token: %w", err)
	}

	return syncToken, nil
}

func (r *MysqlRepository) CreateSyncHistory(ctx context.Context, t *testing.T,
	calendarID valueobject.CalendarID, syncTime time.Time,
	nextSyncToken string, updatedEventCount int) error {
	t.Helper()

	err := createSyncHistory(ctx, r.db, calendarID, syncTime, nextSyncToken, updatedEventCount)
	if err != nil {
		return fmt.Errorf("fail to create sync history: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) CreateSyncHistory(
	ctx context.Context, calendarID valueobject.CalendarID, syncTime time.Time,
	nextSyncToken string, updatedEventCount int) error {

	err := createSyncHistory(ctx, tx.tx, calendarID, syncTime, nextSyncToken, updatedEventCount)
	if err != nil {
		return fmt.Errorf("fail to create sync history: %w", err)
	}

	return nil
}

func createSyncHistory(
	ctx context.Context, db database, calendarID valueobject.CalendarID, syncTime time.Time,
	nextSyncToken string, updatedEventCount int) error {

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO sync_histories "+
			"(calendar_id, sync_time, next_sync_token, updated_event_count) "+
			"VALUES (?, ?, ?, ?)",
		calendarID, syncTime, nextSyncToken, updatedEventCount)

	if err != nil {
		return fmt.Errorf("fail to insert sync history: %w", err)
	}

	return nil
}

func (r *MysqlRepository) DeleteAllSyncHistoriesForMain(ctx context.Context, m *testing.M) (updatedCount int, err error) {
	updatedCount, err = r.deleteAllSyncHistories(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync histories: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) DeleteAllSyncHistories(ctx context.Context, t *testing.T) (updatedCount int, err error) {
	t.Helper()

	updatedCount, err = r.deleteAllSyncHistories(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync histories: %w", err)
	}

	return updatedCount, nil
}

func (r *MysqlRepository) deleteAllSyncHistories(ctx context.Context) (updatedCount int, err error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sync_histories")
	if err != nil {
		return 0, fmt.Errorf("fail to delete all sync histories: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to get affected rows: %w", err)
	}
	updatedCount = int(affectedRows)

	return updatedCount, nil
}
