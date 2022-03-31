package debug_plugin

import (
	"fmt"
	"github.com/pubgo/lava/service"

	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugin"
)

func Enable(srv service.Service) {
	srv.RegisterApp("/debug", debug.App())
	var openWeb bool

	srv.Plugin(&plugin.Base{
		Name:        "debug",
		CfgNotCheck: true,
		OnInit: func(p plugin.Process) {
			p.AfterStart(func() {
				if !openWeb {
					return
				}

				syncx.GoSafe(func() {
					logutil.ErrRecord(zap.L(),
						browser.OpenURL(fmt.Sprintf("http://%s:%d/debug", netutil.GetLocalIP(), srv.Options().Port)))
				})
			})
		},
		OnFlags: func() typex.Flags {
			return typex.Flags{
				&cli.BoolFlag{
					Name:        "debug.web",
					Value:       openWeb,
					Destination: &openWeb,
					Usage:       "open web browser",
				},
			}
		},
	})
}
