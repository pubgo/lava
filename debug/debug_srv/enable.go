package debug_srv

import (
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/service"
)

func Enable(app service.App) {
	app.RegApp("/debug", debug.App())
}
