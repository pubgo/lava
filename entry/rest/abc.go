package rest

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
)

type Service interface {
	Router() func(r Router)
}

type Handler = fiber.Handler
type Router = fiber.Router
type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Plugin(plugins ...plugin.Plugin)
	Register(srv Service, handlers ...Handler)
}
