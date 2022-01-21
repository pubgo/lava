package plugin

import (
	"container/heap"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

type priorityQueue []item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(item))
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)

	if n == 0 {
		return nil
	}

	val := old[n-1]
	*pq = old[0 : n-1]
	return val
}

const defaultPriority = uint(1000)

var plugins = make(map[string]Plugin)

type item struct {
	Name     string
	Priority uint
}

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

// Get 更具名字获取插件
func Get(name string) Plugin { return plugins[name] }

// Middleware 简化Register的注册方法
func Middleware(name string, middleware types.Middleware, priority ...uint) {
	Register(&Base{Name: name, OnMiddleware: middleware}, priority...)
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
		p = p + priority[0]
	}

	plugins[pg.ID()] = pg
	heap.Push(pluginKeys, item{Name: pg.ID(), Priority: p})
}
