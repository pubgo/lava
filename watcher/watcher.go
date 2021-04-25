package watcher

import (
	"context"
	"strings"
	"sync"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var defaultWatcher Watcher
var callbacks types.SMap
var mux sync.Mutex

func Init() (err error) {
	defer xerror.RespErr(&err)

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

		// 获取远程配置
		xerror.Panic(defaultWatcher.GetCallback(context.Background(), name, func(resp *Response) {
			config.GetCfg().Set(KeyToDot(resp.Key), string(resp.Value))
		}))

		// 远程配置watch
		_ = fx.Go(func(ctx context.Context) {
			for resp := range defaultWatcher.Watch(ctx, name) {
				onWatch(resp)
			}
		})
	}

	return
}

func Watch(name string, h CallBack) {
	xerror.Assert(name == "" || h == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "callback %s already exists", name)
	callbacks.Set(name, h)
}

func onWatch(resp *Response) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error("onWatch error", xlog.Any("err", err))
	})

	// 以name为前缀的所有的callbacks
	xerror.Panic(callbacks.Each(func(k string, cb CallBack) {
		mux.Lock()
		defer mux.Unlock()

		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("watch callback error", xlog.Any("err", err))
		})

		key := KeyToDot(resp.Key)

		// 检查是否是以name为前缀, `.`是连接符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 获取数据, 并且更新全局配置
		cfg := config.GetCfg()
		resp.OnDelete(func() { cfg.Set(key, "") })
		resp.OnPut(func() { cfg.Set(key, string(resp.Value)) })

		// 执行watch callback
		var name = KeyToDot(strings.TrimPrefix(key, k))
		xerror.PanicF(cb(name, resp), "event: %#v", *resp)
	}))
}
