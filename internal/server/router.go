package server

import (
	"github.com/gofiber/fiber/v2"
	"server/internal/service/demo"
)

type router struct {
	fiber.Router
	config Config
}

func (r router) setup() {
	r.demo()
	// TODO: More modules here...
}

func (r router) demo() {
	srv := demo.New()

	g := r.Group("demo")
	g.Get("hello", handler(srv.Hello))
	// TODO: More actions here...
}
