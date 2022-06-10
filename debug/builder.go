package debug

import (
	"github.com/pubgo/lava/core/router"
)

func init() {
	router.Register(func(app *router.App) {
		app.Mount("/debug", App())
	})
}
