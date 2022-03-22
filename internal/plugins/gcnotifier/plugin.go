package gcnotifier

import (
	"context"

	"github.com/CAFxX/gcnotifier"

	"github.com/pubgo/lava/logging"
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
	return

	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			p.AfterStop(syncx.GoCtx(func(ctx context.Context) {
				var gc = gcnotifier.New()
				defer gc.Close()

				var gcCh = gc.AfterGC()
				for {
					select {
					case <-gcCh:
						// TODO hook
						logs.L().Info("gc notify")
					case <-ctx.Done():
						return
					}
				}
			}))
		},
	})
}
