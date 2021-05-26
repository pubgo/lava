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

var projectPrefix = make(map[string]struct{})

// InitProjectPrefix 默认的项目前缀是本项目, 可以通过设置, 让plugin前缀支持其他项目
func InitProjectPrefix(projects ...string) {
	for i := range projects {
		projectPrefix[projects[i]] = struct{}{}
	}
}

func GetProjectPrefix() []string {
	var keys = make([]string, 0, len(projectPrefix))
	for k := range projectPrefix {
		keys = append(keys, k)
	}
	return keys
}
