package mysql

import (
	"context"
	"fmt"

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
			"WHERE calendar_id = ? AND status != 'cancelled' "+
			"ORDER BY id",
		calendarID)
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

func (tx *mysqlTransaction) SyncRecurringEvents(ctx context.Context,
	recurringEvent entity.RecurringEvent, events []entity.Event) (updatedCount int, err error) {
	// TODO: Implement recurring event syncing
	return 0, nil
}
