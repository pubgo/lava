package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/entry"
)

type Handler = fiber.Handler
type Router = fiber.Router
type Opts struct{}
type Opt func(opts *Opts)
type Entry interface {
	entry.Entry
	Use(handler ...Handler)
	Router(fn func(r Router))
}
