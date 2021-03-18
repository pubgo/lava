package watcher

import (
	"strings"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var defaultWatcher Watcher
var callbacks types.SMap

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
		mu.Lock()
		defer mu.Unlock()

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
