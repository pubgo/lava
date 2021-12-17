package gcnotifier

import (
	"github.com/CAFxX/gcnotifier"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/syncx"
	"github.com/pubgo/lava/runenv"
)

var Name = "gc"
var logs = logz.Component(Name)

func init() {
	if runenv.IsProd() || runenv.IsRelease() {
		return
	}

	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			syncx.GoSafe(func() {
				var gc = gcnotifier.New()
				defer gc.Close()

				for range gc.AfterGC() {
					logs.Infow("gc notify")
				}
			})
		},
	})
}
