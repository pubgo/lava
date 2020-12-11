package golug_entry_http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_entry_grpc"
)

type Entry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...golug_entry_grpc.Option)
	Use(handler ...fiber.Handler)
	Router(prefix string, fn func(r fiber.Router))
}
