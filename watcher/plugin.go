package watcher

import (
	"context"
	"sync"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

var mu = new(sync.Mutex)

func init() {
	config.On(func(_ *config.Config) {
		defer xerror.RespExit()

		var cfg = GetDefaultCfg()
		if !config.Decode(Name, &cfg) {
			return
		}

		defaultWatcher = xerror.PanicErr(cfg.Build()).(Watcher)

		// 获取所有watch的项目
		projects := cfg.Projects
		if !strutil.Contains(projects, config.Project) {
			projects = append(projects, config.Project)
		}

		// 项目prefix
		for i := range projects {
			var name = projects[i]
			_ = fx.Go(func(ctx context.Context) {
				for resp := range defaultWatcher.Watch(ctx, name) {
					onWatch(resp)
				}
			})
		}
	})
}
