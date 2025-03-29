package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"

	"github.com/takuoki/google-calendar-sync/api/domain"
	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

var refreshTokenCache = service.NewInMemoryCache[valueobject.CalendarID, string]()

func (r *MysqlRepository) GetCalendar(ctx context.Context, calendarID valueobject.CalendarID) (*entity.Calendar, error) {

	var calendar entity.Calendar
	var refreshToken sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, refresh_token FROM calendars WHERE id = ?",
		calendarID,
	).Scan(&calendar.ID, &calendar.Name, &refreshToken)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.CalendarNotFoundError
		}
		return nil, fmt.Errorf("fail to select calendar: %w", err)
	}

	if refreshToken.Valid {
		if r.cryptService != nil {
			decrypted, err := r.cryptService.Decrypt(refreshToken.String)
			if err != nil {
				return nil, fmt.Errorf("fail to decrypt refresh token: %w", err)
			}
			calendar.RefreshToken = &decrypted
		} else {
			calendar.RefreshToken = &refreshToken.String
		}
	}

	if calendar.RefreshToken != nil {
		refreshTokenCache.Set(calendar.ID, *calendar.RefreshToken)
	}

	return &calendar, nil
}

func (r *MysqlRepository) ListCalendars(ctx context.Context) ([]entity.Calendar, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, refresh_token FROM calendars",
	)
	if err != nil {
		return nil, fmt.Errorf("fail to select calendars: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Errorf(ctx, "fail to close rows: %s", closeErr)
		}
	}()

	var calendars []entity.Calendar
	for rows.Next() {
		var calendar entity.Calendar
		var refreshToken sql.NullString

		if err := rows.Scan(&calendar.ID, &calendar.Name, &refreshToken); err != nil {
			return nil, fmt.Errorf("fail to scan calendar: %w", err)
		}

		if refreshToken.Valid {
			if r.cryptService != nil {
				decrypted, err := r.cryptService.Decrypt(refreshToken.String)
				if err != nil {
					return nil, fmt.Errorf("fail to decrypt refresh token: %w", err)
				}
				calendar.RefreshToken = &decrypted
			} else {
				calendar.RefreshToken = &refreshToken.String
			}
		}

		calendars = append(calendars, calendar)

		if calendar.RefreshToken != nil {
			refreshTokenCache.Set(calendar.ID, *calendar.RefreshToken)
		}
	}

	return calendars, nil
}

func (r *MysqlRepository) GetRefreshToken(ctx context.Context, calendarID valueobject.CalendarID) (string, error) {
	if token, ok := refreshTokenCache.Get(calendarID); ok {
		return token, nil
	}

	r.logger.Debug(ctx, "get refresh token from database")

	calendar, err := r.GetCalendar(ctx, calendarID)
	if err != nil {
		return "", fmt.Errorf("fail to get calendar: %w", err)
	}

	if calendar.RefreshToken == nil {
		return "", errors.New("calendar does not have refresh token")
	}

	refreshTokenCache.Set(calendar.ID, *calendar.RefreshToken)
	return *calendar.RefreshToken, nil
}

func (r *MysqlRepository) CreateCalendar(ctx context.Context, t *testing.T, calendar entity.Calendar) error {
	t.Helper()

	err := createCalendar(ctx, r.db, calendar.ID, calendar.Name, calendar.RefreshToken)
	if err != nil {
		return fmt.Errorf("fail to create calendar: %w", err)
	}

	return nil
}

func (tx *mysqlTransaction) CreateCalendar(ctx context.Context, calendar entity.Calendar) error {

	refreshToken := calendar.RefreshToken
	if tx.cryptService != nil && refreshToken != nil {
		encrypted, err := tx.cryptService.Encrypt(*refreshToken)
		if err != nil {
			return fmt.Errorf("fail to encrypt refresh token: %w", err)
		}
		refreshToken = &encrypted
	}

	err := createCalendar(ctx, tx.tx, calendar.ID, calendar.Name, refreshToken)
	if err != nil {
		return fmt.Errorf("fail to create calendar: %w", err)
	}

	if calendar.RefreshToken != nil {
		refreshTokenCache.Set(calendar.ID, *calendar.RefreshToken)
	}

	return nil
}

func createCalendar(ctx context.Context, db database,
	calendarID valueobject.CalendarID, name string, refreshToken *string) error {

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO calendars (id, name, refresh_token) VALUES (?, ?, ?)",
		calendarID, name, refreshToken,
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
