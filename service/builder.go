package service

import (
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(func(app App) {
		app.RegApp("/debug", debug.App())
	})
}
