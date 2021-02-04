package golug_watcher

import (
	"strings"

	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var callbacks golug_types.SyncMap

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
			if strings.HasPrefix(name+".", k.(string)+".") {
				cbs = append(cbs, func(event *Response) error {
					// 获取数据, 并且更新全局配置
					golug_config.GetCfg().Set(KeyToDot(event.Key), string(event.Value))

					return value.(CallBack)(event)
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

	cbs = []CallBack{func(event *Response) error {
		// 获取数据, 并且更新全局配置
		golug_config.GetCfg().Set(KeyToDot(event.Key), string(event.Value))

		return val.(CallBack)(event)
	}}
	return
}
