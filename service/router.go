package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

func init() {
	dix.Provider(func() Router { return func(app *fiber.App) {} })
	dix.Provider(func(routers []Router) *fiber.App {
		var app = fiber.New()
		for i := range routers {
			routers[i](app)
		}
		return app
	})
}

type Router func(app *fiber.App)

type WebHandler interface {
	Router(r fiber.Router)
}
