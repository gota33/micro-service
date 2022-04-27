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
	return entity.Dao[Entity]{
		DB:           db,
		Table:        "item",
		AllFields:    "id, title, price, num, create_time",
		InsertFields: "title, price, num",
		ScanAllFields: func(row entity.Scanner) (e Entity, err error) {
			err = row.Scan(&e.ID, &e.Title, &e.Price, &e.Num, &e.CreateTime)
			return
		},
	}
}
