package debug_srv

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/service"
)

func Enable(srv service.Service) {
	srv.RegApp("/debug", debug.App())

	var openWeb bool
	srv.Flags(&cli.BoolFlag{
		Name:        "debug.web",
		Value:       openWeb,
		Destination: &openWeb,
		Usage:       "open web browser",
	})

	srv.AfterStarts(func() {
		if !openWeb {
			return
		}

		syncx.GoSafe(func() {
			logutil.ErrRecord(zap.L(),
				browser.OpenURL(fmt.Sprintf("http://%s:%d/debug", netutil.GetLocalIP(), srv.Options().Port)))
		})
	})
}
