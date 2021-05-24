package watcher

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var defaultWatcher Watcher
var callbacks typex.SMap
var mux sync.Mutex

func Init() (err error) {
	defer xerror.RespErr(&err)

	var cfg = GetDefaultCfg()
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
		xerror.Panic(defaultWatcher.GetCallback(context.Background(), name, func(resp *Response) {
			var dt interface{}
			xerror.Panic(json.Unmarshal(resp.Value, &dt))
			config.GetCfg().Set(KeyToDot(resp.Key), dt)
		}))

		// 远程配置watch
		defaultWatcher.WatchCallback(context.Background(), name, onWatch)
	}

	return
}

func Watch(name string, cb CallBack) {
	xerror.Assert(name == "" || cb == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "callback %s already exists", name)
	callbacks.Set(name, cb)
}

func onWatch(resp *Response) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error("onWatch error", xlog.Any("err", err))
	})

	resp.OnPut(func() {
		var dt interface{}
		xerror.Panic(jsonx.Unmarshal(resp.Value, &dt))
		config.GetCfg().Set(KeyToDot(resp.Key), dt)
	})

	resp.OnDelete(func() {
		config.GetCfg().Set(KeyToDot(resp.Key), nil)
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

		// 执行watch callback
		var name = KeyToDot(strings.TrimPrefix(key, k))
		xerror.PanicF(cb(name, resp), "event: %#v", *resp)
	}))
}
