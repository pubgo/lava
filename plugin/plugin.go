package plugin

import (
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

var plugins = make(map[string]Plugin)

// pluginKeys 插件key列表, 用来保存插件的注册顺序, 依赖顺序等
var pluginKeys []string

// All 获取所有的插件
func All() []Plugin {
	var pluginList []Plugin
	for _, key := range pluginKeys {
		pluginList = append(pluginList, plugins[key])
	}
	return pluginList
}

// Get 更具名字获取插件
func Get(name string) Plugin { return plugins[name] }

// Middleware 简化Register的注册方法
func Middleware(name string, middleware types.Middleware) {
	Register(&Base{Name: name, OnMiddleware: middleware})
}

// Register 插件注册
func Register(pg Plugin) {
	defer xerror.RespExit("register plugin error")

	xerror.Assert(pg == nil, "plugin[pg] is nil")
	xerror.Assert(pg.ID() == "", "plugin name is null")
	xerror.Assert(plugins[pg.ID()] != nil, "plugin [%s] already exists", pg.ID())

	pluginKeys = append(pluginKeys, pg.ID())
	plugins[pg.ID()] = pg
}
