package watcher

import (
	"strings"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var callbacks types.SMap

func Watch(name string, h CallBack) {
	xerror.Assert(name == "" || h == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "[callback] %s already exists", name)

	callbacks.Set(name, h)
}

func GetWatch(name string, opts ...Option) (cbs []CallBack) {
	var wOpts Options
	for i := range opts {
		opts[i](&wOpts)
	}

	// 以name为前缀的所有的callbacks
	if wOpts.prefix {
		callbacks.Range(func(k, value interface{}) bool {
			// 检查是否是以name为前缀
			// `dot`是连接符
			if strings.HasPrefix(name+".", k.(string)+".") {
				bc := value.(CallBack)
				cbs = append(cbs, func(event *Response) error {
					// 获取数据, 并且更新全局配置
					cfg := config.GetCfg()
					event.OnDelete(func() {
						cfg.Set(KeyToDot(event.Key), "")
					})

					event.OnPut(func() {
						cfg.Set(KeyToDot(event.Key), string(event.Value))
					})

					return bc(event)
				})
			}
			return true
		})
	}

	if len(cbs) == 0 {
		return
	}

	val, ok := callbacks.Load(name)
	if !ok {
		return
	}

	bc := val.(CallBack)
	cbs = []CallBack{func(event *Response) error {
		// 获取数据, 并且更新全局配置
		cfg := config.GetCfg()
		event.OnDelete(func() {
			cfg.Set(KeyToDot(event.Key), "")
		})

		event.OnPut(func() {
			cfg.Set(KeyToDot(event.Key), string(event.Value))
		})

		return bc(event)
	}}
	return
}
