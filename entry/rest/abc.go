package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/entry"
)

type Handler = fiber.Handler
type Router = fiber.Router
type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Use(handler ...Handler)
	Router(fn func(r Router))
}
