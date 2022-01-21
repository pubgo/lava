package pluginInter

import (
	"container/heap"

	"github.com/pubgo/xerror"
)

var plugins = make(map[string]Plugin)

// pluginKeys 插件key列表, 用来保存插件的注册顺序, 依赖顺序等
var pluginKeys = new(priorityQueue)

// All 获取所有的插件
func All() []Plugin {
	var pluginList []Plugin
	for _, key := range *pluginKeys {
		pluginList = append(pluginList, plugins[key.Name])
	}
	return pluginList
}

// Get 根据名字获取插件
func Get(name string) Plugin { return plugins[name] }

// Register 插件注册
//	priority: 优先级
func Register(pg Plugin, priority ...uint) {
	defer xerror.RespExit("register plugin error")

	xerror.Assert(pg == nil, "plugin[pg] is nil")
	xerror.Assert(pg.ID() == "", "plugin name is null")
	xerror.Assert(plugins[pg.ID()] != nil, "plugin [%s] already exists", pg.ID())

	var p = uint(0)
	if len(priority) != 0 {
		p = priority[0]
	}

	plugins[pg.ID()] = pg
	heap.Push(pluginKeys, item{Name: pg.ID(), Priority: p})
}
