package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

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

func (r *mysqlRepository) ListCalendars(ctx context.Context) ([]entity.Calendar, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name FROM calendars",
	)
	if err != nil {
		return nil, fmt.Errorf("fail to select calendars: %w", err)
	}
	defer rows.Close()

	var calendars []entity.Calendar
	for rows.Next() {
		var calendar entity.Calendar
		if err := rows.Scan(&calendar.ID, &calendar.Name); err != nil {
			return nil, fmt.Errorf("fail to scan calendar: %w", err)
		}
		calendars = append(calendars, calendar)
	}

	return calendars, nil
}

func (tx *mysqlTransaction) CreateCalendar(ctx context.Context, calendar entity.Calendar) error {
	_, err := tx.tx.ExecContext(
		ctx,
		"INSERT INTO calendars (id, name) VALUES (?, ?)",
		calendar.ID, calendar.Name,
	)

	if err != nil {
		const duplicateEntryCode = 1062
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == duplicateEntryCode {
			return domain.CalendarAlreadyExistError
		}
		return fmt.Errorf("fail to insert calendar: %w", err)
	}

	return nil
}
