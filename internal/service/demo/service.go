package demo

import (
	"context"

	"server/internal/service/auth"
)

type Service struct{}

func New() Service {
	return Service{}
}

type HelloRequest struct {
	Name string `query:"name" validate:"required,alphaunicode"`
}

type HelloResponse struct {
	Hello string `json:"hello"`
}

func (srv Service) Hello(ctx context.Context, req HelloRequest) (res HelloResponse, err error) {
	var user auth.User
	if err = user.FromContext(ctx); err != nil {
		return
	}

	res = HelloResponse{Hello: req.Name}
	return
}
