package debug_plugin

import (
	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/service/service_type"
)

func Enable(srv service_type.Service) {
	srv.Debug().Mount("/debug", debug.App())

	var openWeb bool

	srv.Flags(
		&cli.BoolFlag{
			Name:        "debug.web",
			Value:       openWeb,
			Destination: &openWeb,
			Usage:       "open web browser with debug",
		},
	)

	srv.AfterStarts(func() {
		if openWeb {
			syncx.GoDelay(func() {
				xerror.Panic(browser.OpenURL("http://localhost:8080/debug"))
			})
		}
	})
}
