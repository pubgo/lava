package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

func Register(prefix string, app *fiber.App) {
	dix.Provider(func() *App {
		return &App{Prefix: prefix, App: app}
	})
}

type App struct {
	Prefix string
	App    *fiber.App
}
