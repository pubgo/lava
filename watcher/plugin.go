package watcher

import (
	"context"
	"sync"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

var mu = new(sync.Mutex)

func init() {
	config.On(func(c *config.Config) {
		defer xerror.RespExit()

		var cfg = GetDefaultCfg()
		config.Decode(Name, &cfg)

		driver := cfg.Driver
		xerror.Assert(driver == "", "watcher driver is null")
		xerror.Assert(!factories.Has(driver), "watcher driver [%s] not found", driver)

		fc := factories.Get(driver).(Factory)
		defaultWatcher = xerror.PanicErr(fc(config.Map(Name))).(Watcher)
		xerror.Assert(defaultWatcher == nil, "watcher driver %s init error", driver)

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
