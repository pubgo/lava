package golug_watcher_etcd

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/plugins/golug_etcd"
)

func init() {
	golug.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		for name, w := range golug_watcher.GetCfg() {
			if w.Driver != golug_etcd.Name {
				continue
			}

			golug_watcher.Register(name, newWatcher(w.Project, golug_etcd.GetClient(w.Name)))
		}
	})
}
