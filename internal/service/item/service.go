package item

import (
	"context"
	"database/sql"
	"net/http"

	"server/internal/service/entity"
)

type Service struct {
	dao dao
}

func New(db *sql.DB) Service {
	return Service{dao: dao{db}}
}

type GetRequest struct {
	ItemID string `param:"itemID"`
}

func (srv Service) Get(ctx context.Context, req GetRequest) (res Entity, err error) {
	return srv.dao.Get(ctx, req.ItemID)
}

type CreateRequest struct {
	Parent     string
	CustomerID string
	Entity
}

func (srv Service) Create(ctx context.Context, req CreateRequest) (res Entity, err error) {
	return srv.dao.Create(ctx, req.Entity)
}

type ListRequest struct {
	entity.ListRequestFragment
}

type ListResponse struct {
	entity.ListResponseFragment
	Items []Entity `json:"items"`
}

func (srv Service) List(ctx context.Context, req ListRequest) (res ListResponse, err error) {
	return srv.dao.List(ctx, req)
}

type UpdateRequest struct {
	ItemID     string           `param:"itemID"`
	UpdateMask entity.FieldMask `json:"updateMask"`
	Item       Entity           `json:"item"`
}

func (srv Service) Update(ctx context.Context, req UpdateRequest) (res Entity, err error) {
	return srv.dao.Update(ctx, req)
}

type DeleteRequest struct {
	ItemID string `param:"itemID"`
}

func (srv Service) Delete(ctx context.Context, req DeleteRequest) (code int, err error) {
	if err = srv.dao.Delete(ctx, req); err == nil {
		code = http.StatusNoContent
	}
	return
}
