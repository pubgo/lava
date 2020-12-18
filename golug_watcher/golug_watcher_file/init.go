package golug_watcher_file

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_watcher"
)

var Name = "file"

func init() {
	// watch file
	golug.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		for name, w := range golug_watcher.GetCfg() {
			if w.Driver != Name {
				continue
			}

			golug_watcher.Register(name, newWatcher())
		}
	})
}
