package gcnotifier

import (
	"context"
	"github.com/pubgo/lava/logging"

	"github.com/CAFxX/gcnotifier"

	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
)

var Name = "gc"
var logs = logging.Component(Name)

func init() {
	if runtime.IsProd() || runtime.IsRelease() {
		return
	}

	plugin.Register(&plugin.Base{
		Name:        Name,
		CfgNotCheck: true,
		Docs:        "Know when GC runs from inside your golang code",
		OnInit: func(p plugin.Process) {
			p.AfterStop(syncx.GoCtx(func(ctx context.Context) {
				var gc = gcnotifier.New()
				defer gc.Close()

				// TODO handler

				for {
					select {
					case <-gc.AfterGC():
						logs.L().Info("gc notify")
					case <-ctx.Done():
						return
					}
				}
			}))
		},
	})
}
