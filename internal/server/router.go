package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"server/internal/service/demo"
)

type router struct {
	fiber.Router
	config   Config
	validate *validator.Validate
}

func (r router) setup() {
	r.demo()
}

func (r router) demo() {
	srv := demo.New()

	g := r.Group("demo")
	g.Get("hello", handler(srv.Hello))
}
