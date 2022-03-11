package gcnotifier

import (
	"github.com/CAFxX/gcnotifier"
	"github.com/pubgo/lava/pkg/syncx"

	"github.com/pubgo/lava/logging"
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
		Name: Name,
		OnInit: func(p plugin.Process) {
			syncx.GoSafe(func() {
				var gc = gcnotifier.New()
				defer gc.Close()

				// TODO hook
				for range gc.AfterGC() {
					logs.L().Info("gc notify")
				}
			})
		},
	})
}
