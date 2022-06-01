package service

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/debug"
)

func init() {
	dix.Register(func(app App) {
		app.RegApp("/debug", debug.App())
	})
}
