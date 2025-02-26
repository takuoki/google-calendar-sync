package mysql

import (
	"context"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (tx *mysqlTransaction) ListActiveChannels(
	ctx context.Context, calendarID valueobject.CalendarID) ([]entity.Channel, error) {

	rows, err := tx.tx.QueryContext(
		ctx,
		"SELECT id, calendar_id, expiration "+
			"FROM channels "+
			"WHERE calendar_id = ? AND expiration > ? AND is_stopped = FALSE "+
			"ORDER BY id FOR UPDATE",
		calendarID, tx.clockService.Now())
	if err != nil {
		// TODO: 空の場合にエラーになるのか要確認
		return nil, fmt.Errorf("fail to select channels: %w", err)
	}
	defer rows.Close()

	var channels []entity.Channel
	for rows.Next() {
		var channel entity.Channel
		err := rows.Scan(&channel.ID, &channel.CalendarID, &channel.Expiration)
		if err != nil {
			return nil, fmt.Errorf("fail to scan row: %w", err)
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (tx *mysqlTransaction) CreateChannel(ctx context.Context, channel entity.Channel) error {

	_, err := tx.tx.ExecContext(
		ctx,
		"INSERT INTO channels "+
			"(id, calendar_id, resource_id, expiration) "+
			"VALUES (?, ?, ?, ?)",
		channel.ID, channel.CalendarID, channel.Expiration)

	if err != nil {
		return fmt.Errorf("fail to insert channel: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) StopChannel(ctx context.Context, channelID valueobject.ChannelID) error {

	_, err := tx.tx.ExecContext(
		ctx,
		"UPDATE channels SET is_stopped = TRUE WHERE id = ?",
		channelID)

	if err != nil {
		return fmt.Errorf("fail to update channel: %w", err)
	}

	return nil
}
