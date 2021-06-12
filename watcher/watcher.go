package watcher

import (
	"bytes"
	"context"
	"strings"

	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var defaultWatcher Watcher = &nullWatcher{}
var callbacks typex.Map

func Init() (err error) {
	defer xerror.RespErr(&err)

	if !config.GetCfg().Decode(Name, &cfg) {
		return
	}

	defaultWatcher = xerror.PanicErr(cfg.Build()).(Watcher)

	// 获取所有watch的项目
	projects := cfg.Projects
	if !strutil.Contains(projects, app.Project) {
		projects = append(projects, app.Project)
	}

	// 项目prefix
	for i := range projects {
		var name = projects[i]

		// 获取远程配置
		xerror.Panic(defaultWatcher.GetCallback(context.Background(), name, func(resp *Response) { onWatch(name, resp) }))

		// 配置远程watch
		defaultWatcher.WatchCallback(context.Background(), name, func(resp *Response) { onWatch(name, resp) })
	}

	vars.Watch(Name+"_callback", func() interface{} {
		var dt []string
		callbacks.Each(func(key string, _ interface{}) { dt = append(dt, key) })
		return dt
	})

	vars.Watch(Name, func() interface{} {
		var dt = make(map[string]string)
		for name, f := range factories {
			dt[name] = stack.Func(f)
		}
		return dt
	})

	return
}

func Watch(name string, cb CallBack) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || cb == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "callback %s already exists", name)
	callbacks.Set(name, cb)
}

func WatchPlugin(project, name string, cb CallBack) {
	name = KeyToDot(project, "plugin", name)
	xerror.Assert(project == "" || name == "" || cb == nil, "[project, name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "callback %s already exists", name)
	callbacks.Set(name, cb)
}

func onWatch(name string, resp *Response) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error("watch handle error",
			xlog.Any("err", err),
			xlog.Any("resp", resp))
	})

	// value是空就skip
	if cfg.SkipNullValue && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = KeyToDot(resp.Key)

	// 把数据设置到全局配置管理中
	// value都必须是kv类型的数据
	var dt = make(map[string]interface{})
	xerror.PanicF(resp.Decode(&dt), "value都必须是kv类型的数据, key:%s, value:%s", resp.Key, resp.Value)

	resp.OnPut(func() {
		if name == app.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), dt)
		} else {
			config.GetCfg().Set(key, dt)
		}
	})

	resp.OnDelete(func() {
		if name == app.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), nil)
		} else {
			config.GetCfg().Set(key, nil)
		}
	})

	// 以name为前缀的所有的callbacks
	callbacks.Each(func(k string, cb interface{}) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("watch callback handle error",
				xlog.String("watch_key", k),
				xlog.Any("err", err))
		})

		// 检查是否是以key为前缀, `.`是连接符和分隔符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 执行watch callback
		var name = strings.Trim(strings.TrimPrefix(key, k), ".")
		xerror.PanicF(cb.(CallBack)(name, resp), "event: %#v", *resp)
	})
}
