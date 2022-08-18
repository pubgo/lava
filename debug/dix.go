package debug

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/service"
)

func init() {
	dix.Provider(func() service.Router {
		return func(app *fiber.App) {
			app.Mount("/debug", App())
		}
	})
}
