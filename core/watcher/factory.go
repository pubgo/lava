package watcher

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/utils"
)

var factories = make(map[string]Factory)
var callbacks = make(map[string][]func(name string, r *Response) error)

// RegisterFactory 注册watcher build factory
func RegisterFactory(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "watcher [name,w] should not be null")
	xerror.Assert(factories[name] != nil, "watcher [%s] already exists", name)
	factories[name] = w
}

func GetFactory(names ...string) Factory {
	val, ok := factories[utils.GetDefault(names...)]
	if !ok {
		return nil
	}

	return val
}

func Watch(name string, callback func(name string, r *Response) error) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || callback == nil, "[name, callback] should not be null")
	callbacks[name] = append(callbacks[name], callback)
}

func GetWatchers() map[string][]func(name string, r *Response) error {
	return callbacks
}
