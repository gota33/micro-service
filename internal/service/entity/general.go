package entity

import (
	"context"
	"database/sql"
	"strconv"
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
	SqlGet        string
	SqlCreate     string
	SqlList       string
	SqlDelete     string
	ScanAllFields func(row Scanner) (e, error)
}

func (d Dao[Entity]) WithDB(db SQLCmd) Dao[Entity] {
	return Dao[Entity]{
		DB:            db,
		Table:         d.Table,
		SqlGet:        d.SqlGet,
		SqlCreate:     d.SqlCreate,
		SqlList:       d.SqlList,
		SqlDelete:     d.SqlDelete,
		ScanAllFields: d.ScanAllFields,
	}
}

func (d Dao[Entity]) Get(ctx context.Context, id string) (e Entity, err error) {
	row := d.DB.QueryRowContext(ctx, d.SqlGet, id)
	return d.ScanAllFields(row)
}

func (d Dao[Entity]) Create(ctx context.Context, e Entity) (next Entity, err error) {
	var (
		tx     SQLCmd
		finish func(error) error
		sr     sql.Result
		id     int64
	)
	if tx, finish, err = BeginTx(ctx, d.DB, nil); err != nil {
		return
	}
	defer func() { err = finish(err) }()

	if sr, err = tx.ExecContext(ctx, d.SqlCreate, e.InsertValues()...); err != nil {
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
	var rows *sql.Rows
	if rows, err = d.DB.QueryContext(ctx, d.SqlList,
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
	size := len(req.UpdateMask.Paths)
	if size == 0 {
		return Validate.Struct(req.Entity)
	}

	names := make([]string, size)
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
		sr  sql.Result
		num int64
	)
	if sr, err = d.DB.ExecContext(ctx, d.SqlDelete, req.ID); err != nil {
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
