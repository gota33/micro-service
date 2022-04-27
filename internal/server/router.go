package server

import (
	"github.com/gofiber/fiber/v2"
	"server/internal/service/demo"
	"server/internal/service/item"
)

type router struct {
	fiber.Router
	config Config
}

func (r router) setup() {
	r.demo()
	r.item()
	// TODO: More modules here...
}

func (r router) demo() {
	srv := demo.New()

	g := r.Group("demo")
	g.Get("hello", handler(srv.Hello))
	// TODO: More actions here...
}

func (r router) item() {
	srv := item.New(r.config.RDS)

	g := r.Group("items")
	g.Post("", handler(srv.Create))
	g.Get("", handler(srv.List))
	g.Get(":itemID", handler(srv.Get))
	g.Patch(":itemID", handler(srv.Update))
	g.Delete(":itemID", handler(srv.Delete))
}
