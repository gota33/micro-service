package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

type FieldMask struct {
	Paths []string `json:"paths,omitempty"`
}

func (mask FieldMask) ToMap(e any) (m map[string]any, err error) {
	var (
		data    []byte
		mFields map[string]any
	)
	if data, err = json.Marshal(e); err != nil {
		return
	}
	if err = json.Unmarshal(data, &mFields); err != nil {
		return
	}

	if size := len(mask.Paths); size > 0 {
		m = make(map[string]any, size)
		for _, path := range mask.Paths {
			if value, ok := mFields[path]; ok {
				m[path] = value
			}
		}
	} else {
		delete(mFields, "id")
		m = mFields
	}
	return
}

type ListRequestFragment struct {
	Parent    string `param:"parent"`
	PageSize  int    `query:"pageSize"`
	PageToken string `query:"pageToken"`
}

type ListResponseFragment struct {
	NextPageToken string `json:"nextPageToken"`
}

type UpdateRequestFragment struct {
	UpdateMask FieldMask `json:"updateMask"`
}

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

func BeginTx(ctx context.Context, db SQLCmd, opts *sql.TxOptions) (tx *sql.Tx, err error) {
	switch db := db.(type) {
	case SQLBeginTx:
		tx, err = db.BeginTx(ctx, opts)
	default:
		err = fmt.Errorf("db should support interface SQLBeginTx")
	}
	return
}

func FinishTx(tx SQLTx, cause error) (err error) {
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

func SQLUpdate(table string, id any, fields map[string]any) (script string, args []any) {
	var sb strings.Builder
	sb.WriteString("update ")
	sb.WriteString(table)
	sb.WriteString(" set ")
	args = append(args, mapJoin(&sb, fields, ", ")...)
	sb.WriteString(" where id = ?")
	args = append(args, id)
	script = sb.String()
	return
}

func mapJoin(sb io.StringWriter, m map[string]any, sep string) (args []any) {
	var flag bool
	for field, value := range m {
		if flag {
			sb.WriteString(sep)
		} else {
			flag = true
		}
		sb.WriteString(field)
		sb.WriteString(" = ?")
		args = append(args, value)
	}
	return
}
