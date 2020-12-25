package golug_watcher_file

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_watcher"
)

var Name = "file"

func init() {
	// watch file
	golug_config.On(func(_ *golug_config.Ctx) {
		for name, w := range golug_watcher.GetCfg() {
			if w.Driver != Name {
				continue
			}

			golug_watcher.Register(name, newWatcher(name))
		}
	})
}
