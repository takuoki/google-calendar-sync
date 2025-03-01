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

func ConnectDB(host, port, user, password, dbname string) (*sql.DB, error) {

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("fail to open database : %w", err)
	}

	return db, nil
}

type mysqlRepository struct {
	db           *sql.DB
	clockService service.Clock
	logger       applog.Logger
}

func NewMysqlRepository(db *sql.DB, clockService service.Clock, logger applog.Logger) repository.DatabaseRepository {
	return &mysqlRepository{
		db:           db,
		clockService: clockService,
		logger:       logger,
	}
}

func (r *mysqlRepository) RunTransaction(ctx context.Context, fn func(ctx context.Context, tx repository.DatabaseTransaction) error) (er error) {
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
		logger:       r.logger,
	})
}

type mysqlTransaction struct {
	tx           *sql.Tx
	clockService service.Clock
	logger       applog.Logger
}
