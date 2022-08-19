package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

type Web struct {
	*fiber.App
}

func init() {
	dix.Provider(func() Router { return func(app *fiber.App) {} })
	dix.Provider(func(routers []Router) *Web {
		var app = fiber.New()
		for i := range routers {
			routers[i](app)
		}
		return &Web{App: app}
	})
}

type Router func(app *fiber.App)

type WebHandler interface {
	Router(r fiber.Router)
}
