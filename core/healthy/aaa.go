package healthy

import (
	"github.com/gofiber/fiber/v3"
)

type Handler func(req fiber.Ctx) error
