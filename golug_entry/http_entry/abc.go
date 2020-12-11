package http_entry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/grpc_entry"
)

type Entry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...grpc_entry.Option)
	Use(handler ...fiber.Handler)
	Router(prefix string, fn func(r fiber.Router))
}
