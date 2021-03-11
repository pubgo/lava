package plugin

const defaultModule = "__default"

var dm = newManager()

func String() string                             { return dm.String() }
func All() map[string][]Plugin                   { return dm.All() }
func List(opts ...ManagerOpt) []Plugin           { return dm.Plugins(opts...) }
func Register(plugin Plugin, opts ...ManagerOpt) { dm.Register(plugin, opts...) }

// IsRegistered check plugin whether registered global.
// Notice plugin is not check whether is nil
func IsRegistered(plugin Plugin, opts ...ManagerOpt) bool { return dm.isRegistered(plugin, opts...) }
