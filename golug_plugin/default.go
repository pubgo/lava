package golug_plugin

import "github.com/pubgo/xerror"

const defaultModule = "__golug"

var defaultManager = newManager()

func String() string {
	return defaultManager.String()
}

func All() map[string][]Plugin {
	return defaultManager.All()
}

// List lists the global plugins
func List(opts ...ManagerOption) []Plugin {
	return defaultManager.Plugins(opts...)
}

// Register registers a global plugins
func Register(plugin Plugin, opts ...ManagerOption) {
	xerror.Panic(defaultManager.Register(plugin, opts...))
}

// IsRegistered check plugin whether registered global.
// Notice plugin is not check whether is nil
func IsRegistered(plugin Plugin, opts ...ManagerOption) bool {
	return defaultManager.isRegistered(plugin, opts...)
}
