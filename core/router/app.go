package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

func Register(fn func(app *App)) { dix.Register(fn) }

func init() {
	dix.Register(func() *App { return &App{App: fiber.New()} })
}

type App struct {
	*fiber.App
}
