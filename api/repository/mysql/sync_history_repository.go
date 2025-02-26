package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *mysqlRepository) GetLatestSyncToken(
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

func (tx *mysqlTransaction) CreateSyncHistory(
	ctx context.Context, calendarID valueobject.CalendarID, syncTime time.Time,
	nextSyncToken string, updatedEventCount int) error {

	_, err := tx.tx.ExecContext(
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
