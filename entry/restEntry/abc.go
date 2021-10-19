package restEntry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/entry"
)

type Handler = fiber.Handler
type Router = fiber.Router
type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Register(srv interface{}, opts ...Opt)
}
type RestRouter interface {
	Router(r Router)
}
