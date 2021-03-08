package watcher

import (
	"context"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/xerror"
)

func init() {
	// 服务启动后, 启动配置监控
	golug_run.AfterStart(func() {
		config.Decode(Name, &cfgList)

		for _, cfg := range cfgList {
			xerror.Assert(Get(cfg.Driver) == nil, "watcher %s not exists", cfg.Driver)

			go func(driver string, project string) {
				w := Get(driver)

				for resp := range w.Watch(context.Background(), project) {
					onWatch(resp)
				}
			}(cfg.Driver, cfg.Project)
		}
	})
}
