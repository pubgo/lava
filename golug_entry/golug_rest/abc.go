package golug_rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_entry"
)

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...Option)
	Use(handler ...fiber.Handler)
	Router(fn func(r fiber.Router))
}
