package item

import (
	"context"
	"database/sql"
	"net/http"

	"server/internal/service/entity"
)

type Service struct {
	dao entity.Dao[Entity]
}

func New(db *sql.DB) Service {
	return Service{dao: newDao(db)}
}

type GetRequest struct {
	ItemID string `param:"itemID"`
}

func (srv Service) Get(ctx context.Context, req GetRequest) (res Entity, err error) {
	return srv.dao.Get(ctx, req.ItemID)
}

func (srv Service) Create(ctx context.Context, entity Entity) (res Entity, err error) {
	return srv.dao.Create(ctx, entity)
}

type ListRequest struct {
	entity.ListRequestFragment
}

type ListResponse struct {
	entity.ListResponseFragment
	Items []Entity `json:"items"`
}

func (srv Service) List(ctx context.Context, req ListRequest) (res ListResponse, err error) {
	// Change default pageSize if needed
	// req.FallbackPageSize = 10

	var raw entity.ListResponse[Entity]
	if raw, err = srv.dao.List(ctx, req); err != nil {
		return
	}

	res.Items = raw.Items
	res.ListResponseFragment = raw.ListResponseFragment
	return
}

type UpdateRequest struct {
	entity.UpdateRequestFragment
	ItemID string `param:"itemID"`
	Item   Entity `json:"item"`
}

func (srv Service) Update(ctx context.Context, req UpdateRequest) (res Entity, err error) {
	uReq := entity.UpdateRequest[Entity]{
		UpdateRequestFragment: req.UpdateRequestFragment,
		ID:                    req.ItemID,
		Entity:                req.Item,
	}
	return srv.dao.Update(ctx, uReq)
}

type DeleteRequest struct {
	ItemID string `param:"itemID"`
}

func (srv Service) Delete(ctx context.Context, req DeleteRequest) (code int, err error) {
	if err = srv.dao.Delete(ctx, req.ItemID); err == nil {
		code = http.StatusNoContent
	}
	return
}
