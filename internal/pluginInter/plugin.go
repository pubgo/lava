package pluginInter

import (
	"container/heap"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
)

var plugins = make(map[string]Plugin)

// pluginKeys 插件key列表, 用来保存插件的注册顺序, 依赖顺序等
var pluginKeys typex.PriorityQueue

// All 获取所有的插件
func All() []Plugin {
	var pluginList = make([]Plugin, len(pluginKeys))
	for _, key := range pluginKeys {
		pluginList = append(pluginList, plugins[key.Value.(string)])
	}
	return pluginList
}

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
	heap.Push(&pluginKeys, &typex.PriorityQueueItem{Priority: int64(p), Value: pg.ID()})
}
