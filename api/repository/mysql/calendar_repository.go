package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

func (r *mysqlRepository) GetCalendar(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Calendar, error) {
	var calendar entity.Calendar

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name FROM calendars WHERE id = ?",
		calendarID,
	).Scan(&calendar.ID, &calendar.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.CalendarNotFoundError
		}
		return nil, fmt.Errorf("fail to select calendar: %w", err)
	}

	return &calendar, nil
}
