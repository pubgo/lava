package watcher

import (
	"context"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	var cfg = GetDefaultCfg()
	config.Decode(Name, &cfg)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(!factories.Has(driver), "watcher driver %s is null", driver)

	fc := factories.Get(driver).(Factory)
	xerror.Assert(fc == nil, "watcher driver %s not found", driver)

	defaultWatcher = xerror.PanicErr(fc(config.Map(Name))).(Watcher)
	xerror.Assert(defaultWatcher == nil, "watcher driver %s init error", driver)

	projects := cfg.Projects
	if !strutil.Contains(projects, config.Project) {
		projects = append(projects, config.Project)
	}

	// 项目prefix
	for i := range projects {
		go func(name string) {
			for resp := range defaultWatcher.Watch(context.Background(), name) {
				onWatch(resp)
			}
		}(projects[i])
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})

	tracelog.Watch(Name+"_watcher_callback", func() interface{} {
		var dt []string
		xerror.Panic(callbacks.Each(func(key string) { dt = append(dt, key) }))
		return dt
	})

	tracelog.Watch(Name+"_watcher", func() interface{} {
		var dt = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, f Factory) {
			dt[name] = stack.Func(f)
		}))
		return dt
	})
}
