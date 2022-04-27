package entity

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gota33/errors"
)

type Entity interface {
	GetID() string
	InsertValues() []any
}

type Dao[e Entity] struct {
	DB            SQLCmd
	Table         string
	AllFields     string
	InsertFields  string
	ScanAllFields func(row Scanner) (e, error)
}

func (d Dao[Entity]) WithDB(db SQLCmd) Dao[Entity] {
	return Dao[Entity]{
		DB:            db,
		Table:         d.Table,
		AllFields:     d.AllFields,
		InsertFields:  d.InsertFields,
		ScanAllFields: d.ScanAllFields,
	}
}

func (d Dao[Entity]) Get(ctx context.Context, id string) (e Entity, err error) {
	script := fmt.Sprintf("select %s from %s where id = ? limit 1", d.AllFields, d.Table)
	row := d.DB.QueryRowContext(ctx, script, id)
	return d.ScanAllFields(row)
}

func (d Dao[Entity]) Create(ctx context.Context, e Entity) (next Entity, err error) {
	var (
		holders = "?" + strings.Repeat(",?", strings.Count(d.InsertFields, ","))
		script  = fmt.Sprintf("insert into %s (%s) values (%s)", d.Table, d.InsertFields, holders)
		tx      SQLCmd
		finish  func(error) error
		sr      sql.Result
		id      int64
	)
	if tx, finish, err = BeginTx(ctx, d.DB, nil); err != nil {
		return
	}
	defer func() { err = finish(err) }()

	if sr, err = tx.ExecContext(ctx, script, e.InsertValues()...); err != nil {
		return
	}
	if id, err = sr.LastInsertId(); err != nil {
		return
	}

	sub := d.WithDB(tx)
	return sub.Get(ctx, strconv.FormatInt(id, 10))
}

type ListRequest interface {
	GetPageSize() int
	GetPageToken() string
}

type ListResponse[e Entity] struct {
	ListResponseFragment
	Items []e
}

func (d Dao[Entity]) List(ctx context.Context, req ListRequest) (res ListResponse[Entity], err error) {
	var (
		script = fmt.Sprintf("select %s from %s where id > ? limit ?", d.AllFields, d.Table)
		rows   *sql.Rows
	)
	if rows, err = d.DB.QueryContext(ctx, script,
		req.GetPageToken(), req.GetPageSize()); err != nil {
		return
	}

	defer CloseRows(rows)

	for rows.Next() {
		var e Entity
		if e, err = d.ScanAllFields(rows); err != nil {
			return
		}
		res.Items = append(res.Items, e)
	}
	if err = rows.Err(); err != nil {
		return
	}

	if size := len(res.Items); size == req.GetPageSize() {
		res.NextPageToken = res.Items[size-1].GetID()
	}
	return
}

type UpdateRequest[E Entity] struct {
	UpdateRequestFragment
	ID     string
	Entity E
}

func (req UpdateRequest[Entity]) validate() (err error) {
	names := make([]string, len(req.UpdateMask.Paths))
	for i, path := range req.UpdateMask.Paths {
		runes := []rune(path)
		runes[0] = unicode.ToUpper(runes[0])
		names[i] = string(runes)
	}
	return Validate.StructPartial(req.Entity, names...)
}

func (d Dao[Entity]) Update(ctx context.Context, req UpdateRequest[Entity]) (res Entity, err error) {
	if err = req.validate(); err != nil {
		return
	}

	var fields map[string]any
	if fields, err = req.UpdateMask.ToMap(req.Entity); err != nil {
		return
	}
	script, args := SQLUpdate(d.Table, req.ID, fields)

	var (
		tx     SQLCmd
		finish func(error) error
	)
	if tx, finish, err = BeginTx(ctx, d.DB, nil); err != nil {
		return
	}
	defer func() { err = finish(err) }()

	if _, err = tx.ExecContext(ctx, script, args...); err != nil {
		return
	}

	sub := d.WithDB(tx)
	return sub.Get(ctx, req.ID)
}

type DeleteRequest struct {
	errors.ResourceInfo
	ID string
}

func (d Dao[Entity]) Delete(ctx context.Context, req DeleteRequest) (err error) {
	var (
		script = fmt.Sprintf("delete from %s where id = ?", d.Table)
		sr     sql.Result
		num    int64
	)
	if sr, err = d.DB.ExecContext(ctx, script, req.ID); err != nil {
		return
	}
	if num, err = sr.RowsAffected(); err != nil {
		return
	}
	if num == 0 {
		err = errors.WithNotFound(errors.NotFound, req.ResourceInfo)
	}
	return
}
