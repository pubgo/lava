package plugin

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/internal/pluginInter"
	"github.com/pubgo/lava/service/service_type"
	"github.com/pubgo/xerror"
)

const defaultPriority = uint(1000)

// All 获取所有的插件
func All() []Plugin { return pluginInter.All() }

func Init(p config_type.Config, plugins ...Plugin) {
	for _, plg := range append(All(), plugins...) {
		xerror.Panic(plg.Init(p))
	}
}

// Middleware 简化Register的注册方法
func Middleware(name string, middleware service_type.Middleware, priority ...uint) {
	Register(&Base{Name: name, OnMiddleware: middleware}, priority...)
}

// Register 插件注册
//	priority: 优先级
func Register(pg Plugin, priority ...uint) {
	var p = defaultPriority
	if len(priority) != 0 {
		p = priority[0]
	}
	pluginInter.Register(pg, p)
}
