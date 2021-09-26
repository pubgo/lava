package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
)

type Handler = fiber.Handler
type Router = fiber.Router
type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Register(srv interface{})
	Plugin(plugins ...plugin.Plugin)
}

func Provider(fn func(r Router)) {
	xerror.Exit(dix.Provider(fn))
}
