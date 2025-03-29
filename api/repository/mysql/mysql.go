package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/takuoki/golib/applog"
	"github.com/takuoki/google-calendar-sync/api/domain/service"
	"github.com/takuoki/google-calendar-sync/api/repository"

	_ "github.com/go-sql-driver/mysql"
)

type database interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func ConnectDB(host, port, user, password, dbname string) (*sql.DB, error) {

	if err := validateArgs(host, user, password, dbname); err != nil {
		return nil, err
	}

	if port == "" {
		port = "3306"
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("fail to open database : %w", err)
	}

	return db, nil
}

func validateArgs(host, user, password, dbname string) error {
	if host == "" {
		return fmt.Errorf("database host is required")
	}
	if user == "" {
		return fmt.Errorf("database user is required")
	}
	if password == "" {
		return fmt.Errorf("database password is required")
	}
	if dbname == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

type MysqlRepository struct {
	db           *sql.DB
	clockService service.Clock
	cryptService service.Crypt
	logger       applog.Logger
}

func NewMysqlRepository(db *sql.DB, clockService service.Clock,
	cryptService service.Crypt, logger applog.Logger) *MysqlRepository {
	return &MysqlRepository{
		db:           db,
		clockService: clockService,
		cryptService: cryptService,
		logger:       logger,
	}
}

func (r *MysqlRepository) RunTransaction(ctx context.Context, fn func(ctx context.Context, tx repository.DatabaseTransaction) error) (er error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rerr := tx.Rollback(); rerr != nil {
				r.logger.Errorf(ctx, "fail to rollback: %v", rerr)
			}
			panic(p)
		}

		if er != nil {
			if rerr := tx.Rollback(); rerr != nil {
				// ここでエラーが発生した場合、戻り値には元のエラーを優先する
				r.logger.Errorf(ctx, "fail to rollback: %v", rerr)
			}
		} else {
			if cerr := tx.Commit(); cerr != nil {
				er = fmt.Errorf("fail to commit: %w", cerr)
			}
		}
	}()

	return fn(ctx, &mysqlTransaction{
		tx:           tx,
		clockService: r.clockService,
		cryptService: r.cryptService,
		logger:       r.logger,
	})
}

type mysqlTransaction struct {
	tx           *sql.Tx
	clockService service.Clock
	cryptService service.Crypt
	logger       applog.Logger
}
