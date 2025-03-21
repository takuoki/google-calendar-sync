package mysql

import (
	"context"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (tx *mysqlTransaction) ListActiveChannelHistoriesWithLock(
	ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Channel, error) {

	rows, err := tx.tx.QueryContext(
		ctx,
		"SELECT calendar_id, start_time, resource_id, expiration "+
			"FROM channel_histories "+
			"WHERE calendar_id = ? AND expiration > ? AND is_stopped = FALSE "+
			"ORDER BY start_time FOR UPDATE",
		calendarID, tx.clockService.Now())
	if err != nil {
		// TODO: レコードを取得できない場合にエラーになるのか要確認
		return nil, fmt.Errorf("fail to select channel history: %w", err)
	}
	defer rows.Close()

	var channels []entity.Channel
	for rows.Next() {
		var channel entity.Channel
		err := rows.Scan(&channel.CalendarID, &channel.StartTime,
			&channel.ResourceID, &channel.Expiration)
		if err != nil {
			return nil, fmt.Errorf("fail to scan row: %w", err)
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (tx *mysqlTransaction) CreateChannelHistory(
	ctx context.Context, channel entity.Channel) error {

	_, err := tx.tx.ExecContext(
		ctx,
		"INSERT INTO channel_histories "+
			"(calendar_id, start_time, resource_id, expiration) "+
			"VALUES (?, ?, ?, ?)",
		channel.CalendarID, channel.StartTime, channel.ResourceID, channel.Expiration)

	if err != nil {
		return fmt.Errorf("fail to insert channel history: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) StopActiveChannels(
	ctx context.Context, calendarID valueobject.CalendarID) error {

	_, err := tx.tx.ExecContext(
		ctx,
		"UPDATE channel_histories SET is_stopped = TRUE "+
			"WHERE calendar_id = ? AND expiration > ? AND is_stopped = FALSE",
		calendarID, tx.clockService.Now())

	if err != nil {
		return fmt.Errorf("fail to update channel history: %w", err)
	}

	return nil
}
