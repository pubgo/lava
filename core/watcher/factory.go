package watcher

import (
	"github.com/pubgo/lava/core/watcher/watcher_type"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/utils"
)

var factories = make(map[string]watcher_type.Factory)
var callbacks = make(map[string][]func(name string, r *watcher_type.Response) error)

// RegisterFactory 注册watcher build factory
func RegisterFactory(name string, w watcher_type.Factory) {
	xerror.Assert(name == "" || w == nil, "watcher [name,w] should not be null")
	xerror.Assert(factories[name] != nil, "watcher [%s] already exists", name)
	factories[name] = w
}

func GetFactory(names ...string) watcher_type.Factory {
	val, ok := factories[utils.GetDefault(names...)]
	if !ok {
		return nil
	}

	return val
}

func Watch(name string, callback func(name string, r *watcher_type.Response) error) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || callback == nil, "[name, callback] should not be null")
	callbacks[name] = append(callbacks[name], callback)
}

func GetWatchers() map[string][]func(name string, r *watcher_type.Response) error {
	return callbacks
}
