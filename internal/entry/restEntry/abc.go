package restEntry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/entry"
)

type Entry interface {
	entry.Entry
	Register(srv Handler)
}

type Handler interface {
	entry.InitHandler
	Router(r fiber.Router)
}
