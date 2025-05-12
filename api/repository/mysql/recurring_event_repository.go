package mysql

import (
	"context"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *MysqlRepository) ListRecurringEvents(ctx context.Context, calendarID valueobject.CalendarID) (
	[]entity.RecurringEvent, error) {
	// TODO: Implement recurring event fetching
	return nil, nil
}

func (tx *mysqlTransaction) SyncRecurringEvents(ctx context.Context,
	recurringEvent entity.RecurringEvent, events []entity.Event) (updatedCount int, err error) {
	// TODO: Implement recurring event syncing
	return 0, nil
}
