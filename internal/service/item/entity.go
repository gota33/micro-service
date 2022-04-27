package item

import (
	"database/sql"
	"strconv"
	"time"

	"server/internal/service/entity"
)

type Entity struct {
	ID         int64     `json:"id,string"`
	Title      string    `json:"title" validate:"required"`
	Price      float64   `json:"price" validate:"required,min=0"`
	Num        int64     `json:"num" validate:"required,min=0"`
	CreateTime time.Time `json:"createTime"`
}

func (e Entity) GetID() string {
	return strconv.FormatInt(e.ID, 10)
}

func (e Entity) InsertValues() []any {
	return []any{e.Title, e.Price, e.Num}
}

func newDao(db *sql.DB) entity.Dao[Entity] {
	const allFields = "id, title, price, num, create_time"
	return entity.Dao[Entity]{
		DB:        db,
		Table:     "item",
		SqlCreate: "insert into item (title, price, num) values (?, ?, ?)",
		SqlGet:    "select " + allFields + " from item where id = ? limit 1",
		SqlList:   "select " + allFields + " from item where id > ? limit ?",
		SqlDelete: "delete from item where id = ?",
		ScanAllFields: func(row entity.Scanner) (e Entity, err error) {
			err = row.Scan(&e.ID, &e.Title, &e.Price, &e.Num, &e.CreateTime)
			return
		},
	}
}
