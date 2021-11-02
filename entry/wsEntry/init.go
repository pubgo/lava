package wsEntry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func init() {
	app := fiber.New()
	app.Get("/ws", websocket.New(nil))
}
