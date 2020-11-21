package golug_entry

import (
	"github.com/gofiber/fiber/v2"
)

type HttpEntry interface {
	Entry
	Use(handler ...fiber.Handler)
	Group(prefix string, fn func(r fiber.Router))
}
