package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
)

type WebApp struct {
	*fiber.App
}

func init() {
	defer recovery.Exit()
	dix.Provider(func() Router { return func(app *fiber.App) {} })
	dix.Provider(func(routers []Router) *WebApp {
		var app = fiber.New()
		for i := range routers {
			routers[i](app)
		}
		return &WebApp{App: app}
	})
}

type Router func(app *fiber.App)

type WebHandler interface {
	Router(r fiber.Router)
}
