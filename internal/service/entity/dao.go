package entity

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Scanner interface {
	Scan(dest ...any) error
}

type SQLCmd interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type SQLBeginTx interface {
	SQLCmd
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SQLTx interface {
	SQLCmd
	Commit() error
	Rollback() error
}

func BeginTx(ctx context.Context, db SQLCmd, opts *sql.TxOptions) (tx SQLCmd, finish func(error) error, err error) {
	switch db := db.(type) {
	case SQLBeginTx:
		tx, err = db.BeginTx(ctx, opts)
		finish = func(cause error) error { return finishTx(tx.(*sql.Tx), cause) }
	default:
		tx = db
		finish = func(cause error) error { return cause }
	}
	return
}

func finishTx(tx SQLTx, cause error) (err error) {
	if err = cause; err == nil {
		return tx.Commit()
	}
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		logrus.WithError(rollbackErr).Warnf("Rollback error")
	}
	return
}

func CloseRows(rows *sql.Rows) {
	if closeErr := rows.Err(); closeErr != nil {
		logrus.WithError(closeErr).Warn("Close rows error")
	}
}
