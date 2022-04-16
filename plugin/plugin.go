package plugin

import (
	"container/heap"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/resource"
)

var plugins = make(map[string]Plugin)

// pluginKeys 插件key列表, 用来保存插件的注册顺序, 依赖顺序等
var pluginKeys typex.PriorityQueue

const defaultPriority = uint(1000)

// RegisterMiddleware 简化Register的注册方法
func RegisterMiddleware(name string, middleware Middleware, priority ...uint) {
	Register(&Base{Name: name, OnMiddleware: middleware, CfgNotCheck: true}, priority...)
}

// RegisterResource 构建资源
func RegisterResource(name string, builder resource.BuilderFactory, priority ...uint) {
	Register(&Base{Name: name, BuilderFactory: builder}, priority...)
}

// RegisterProcess process注册
func RegisterProcess(name string, p func(p Process), priority ...uint) {
	Register(&Base{Name: name, OnInit: p, CfgNotCheck: true}, priority...)
}

// All 获取所有的插件
func All() []Plugin {
	var pluginList = make([]Plugin, len(pluginKeys))
	for i, key := range pluginKeys {
		pluginList[i] = plugins[key.Value.(string)]
	}
	return pluginList
}

// Get 获取插件
func Get(name string) Plugin {
	return plugins[name]
}

// Register 插件注册
//	priority: 优先级
func Register(pg Plugin, priority ...uint) {
	defer xerror.RespExit("register plugin error")

	xerror.Assert(pg == nil, "plugin[pg] is nil")
	xerror.Assert(pg.ID() == "", "plugin name is null")
	xerror.Assert(plugins[pg.ID()] != nil, "plugin [%s] already exists", pg.ID())

	var p = defaultPriority
	if len(priority) != 0 {
		p = priority[0]
	}

	plugins[pg.ID()] = pg
	heap.Push(&pluginKeys, &typex.PriorityQueueItem{Priority: int64(p), Value: pg.ID()})
}
