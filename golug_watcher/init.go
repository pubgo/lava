package golug_watcher

import (
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/xerror"
)

func init() {
	// 服务启动后, 启动配置监控
	golug_run.AfterStart(func() {
		for _, w := range List() {
			xerror.ExitF(w.Start(), "watcher %s start error", w.Name())
		}
	})

	// 停止服务之后, 关闭配置的监控
	golug_run.BeforeStop(func() {
		for _, w := range List() {
			xerror.ExitF(w.Close(), "watcher %s close error", w.Name())
		}
	})
}
