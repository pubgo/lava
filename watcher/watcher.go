package watcher

import (
	"strings"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var callbacks types.SMap

func Watch(name string, h CallBack) {
	xerror.Assert(name == "" || h == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "[callback] %s already exists", name)

	callbacks.Set(name, h)
}

func onWatch(resp *Response) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error("", xlog.Any("err", err))
	})

	// 以name为前缀的所有的callbacks
	xerror.Panic(callbacks.Each(func(k string, bc CallBack) {
		// 检查是否是以name为前缀, `dot`是连接符
		if !strings.HasPrefix(resp.Key+".", k+".") {
			return
		}

		// 获取数据, 并且更新全局配置
		cfg := config.GetCfg()
		resp.OnDelete(func() { cfg.Set(KeyToDot(resp.Key), "") })
		resp.OnPut(func() { cfg.Set(KeyToDot(resp.Key), string(resp.Value)) })
		xerror.Panic(bc(resp))
	}))
}
