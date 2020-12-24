package golug_watcher

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
)

func init() {
	var watchers []Watcher
	// 服务启动后, 启动配置监控
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		for _, w := range List() {
			wc := w()
			if wc == nil {
				continue
			}

			watchers = append(watchers, wc)
			xerror.ExitF(wc.Start(), wc.Name())
		}
	}))

	// 停止服务之后, 关闭配置的监控
	xerror.Exit(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) {
		for _, w := range watchers {
			xerror.ExitF(w.Close(), w.Name())
		}
	}))
}
