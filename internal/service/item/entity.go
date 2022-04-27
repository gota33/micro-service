package item

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/gota33/errors"
	. "server/internal/service/entity"
)

const (
	allFields = "id, title, price, num, create_time"
)

type Entity struct {
	ID         int64     `json:"id,string"`
	Title      string    `json:"title"`
	Price      float64   `json:"price"`
	Num        int64     `json:"num"`
	CreateTime time.Time `json:"createTime"`
}

func (e *Entity) scanAllFields(row Scanner) error {
	return row.Scan(&e.ID, &e.Title, &e.Price, &e.Num, &e.CreateTime)
}

type dao struct {
	db SQLCmd
}

func (d dao) Get(ctx context.Context, id string) (e Entity, err error) {
	const script = "select " + allFields + " from item where id = ? limit 1"
	row := d.db.QueryRowContext(ctx, script, id)
	err = e.scanAllFields(row)
	return
}

func (d dao) Create(ctx context.Context, e Entity) (next Entity, err error) {
	const script = "insert into item (title, price, num) values (?, ?, ?)"
	var (
		tx SQLTx
		sr sql.Result
		id int64
	)
	if tx, err = BeginTx(ctx, d.db, nil); err != nil {
		return
	}
	defer func() { err = FinishTx(tx, err) }()

	sub := dao{db: tx}

	if sr, err = tx.ExecContext(ctx, script, e.Title, e.Price, e.Num); err != nil {
		return
	}
	if id, err = sr.LastInsertId(); err != nil {
		return
	}
	if next, err = sub.Get(ctx, strconv.FormatInt(id, 10)); err != nil {
		return
	}
	return
}

func (d dao) List(ctx context.Context, req ListRequest) (res ListResponse, err error) {
	const script = "select " + allFields + " from item where id > ? limit ?"

	if req.PageSize == 0 {
		req.PageSize = 20
	}

	if req.PageToken == "" {
		req.PageToken = "0"
	}

	var rows *sql.Rows
	if rows, err = d.db.QueryContext(ctx, script, req.PageToken, req.PageSize); err != nil {
		return
	}

	defer CloseRows(rows)

	for rows.Next() {
		var e Entity
		if err = e.scanAllFields(rows); err != nil {
			return
		}
		res.Items = append(res.Items, e)
	}
	if err = rows.Err(); err != nil {
		return
	}

	if size := len(res.Items); size == req.PageSize {
		res.NextPageToken = strconv.FormatInt(res.Items[size-1].ID, 10)
	}
	return
}

func (d dao) Update(ctx context.Context, req UpdateRequest) (res Entity, err error) {
	var fields map[string]any
	if fields, err = req.UpdateMask.ToMap(req.Item); err != nil {
		return
	}
	script, args := SQLUpdate("item", req.ItemID, fields)

	var tx SQLTx
	if tx, err = BeginTx(ctx, d.db, nil); err != nil {
		return
	}
	defer func() { err = FinishTx(tx, err) }()

	sub := dao{db: tx}

	if _, err = tx.ExecContext(ctx, script, args...); err != nil {
		return
	}
	if res, err = sub.Get(ctx, req.ItemID); err != nil {
		return
	}
	return
}

func (d dao) Delete(ctx context.Context, req DeleteRequest) (err error) {
	const script = "delete from item where id = ?"
	var (
		sr  sql.Result
		num int64
	)
	if sr, err = d.db.ExecContext(ctx, script, req.ItemID); err != nil {
		return
	}
	if num, err = sr.RowsAffected(); err != nil {
		return
	}
	if num == 0 {
		err = errors.WithNotFound(errors.NotFound, errors.ResourceInfo{
			ResourceType: "item",
			ResourceName: "items/" + req.ItemID,
		})
	}
	return
}

/*
func (d dao) OldCreate(ctx context.Context, e *Entity) (err error) {
	var (
		tx *sql.Tx
		sr sql.Result
	)
	if tx, err = d.db.BeginTx(ctx, &sql.TxOptions{}); err != nil {
		return
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logrus.WithError(rollbackErr).Warnf("Rollback error")
			} else {
				err = tx.Commit()
			}
		}
	}()

	if sr, err = tx.ExecContext(ctx, sqlCreate, e.Title, e.Price, e.Num); err != nil {
		return
	}
	if e.ID, err = sr.LastInsertId(); err != nil {
		return
	}
	row := tx.QueryRowContext(ctx, sqlGet)
	err = e.scanAllFields(row)
	return
}
*/
