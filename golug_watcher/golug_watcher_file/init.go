package golug_watcher_file

import (
	"github.com/pubgo/golug/golug_watcher"
)

var Name = "file"

func init() {
	golug_watcher.Register(Name, func() golug_watcher.Watcher {
		// watch file
		for name, w := range golug_watcher.GetCfg() {
			if w.Driver != Name {
				continue
			}

			return newWatcher(name)
		}
		return nil
	})
}
