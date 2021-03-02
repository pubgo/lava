package golug_rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/entry"
)

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	entry.Entry
	Use(handler ...fiber.Handler)
	Router(fn func(r fiber.Router))
}
