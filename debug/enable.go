package debug

import (
	"fmt"
	"github.com/pubgo/lava/service/service_type"

	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/runtime"
)

func Enable(srv service_type.Service) {
	srv.Debug().Mount("/", app)

	var openWeb bool

	srv.Flags(
		&cli.BoolFlag{
			Name:        "debug.web",
			Value:       openWeb,
			Destination: &openWeb,
			Usage:       "open web browser with debug",
		},
		&cli.StringFlag{
			Name:        "debug.addr",
			Destination: &runtime.DebugAddr,
			Usage:       "debug server http address",
			Value:       runtime.DebugAddr,
			EnvVars:     env.KeyOf("lava-debug-addr"),
		},
	)

	srv.AfterStarts(func() {
		if openWeb {
			syncx.GoDelay(func() {
				xerror.Panic(browser.OpenURL(fmt.Sprintf("http://localhost:%d/debug", srv.Options().Port)))
			})
		}
	})

}
