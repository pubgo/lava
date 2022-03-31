package plugin

import (
	"encoding/json"
	"github.com/pubgo/lava/core/healthy"
	"github.com/pubgo/lava/core/watcher"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/abc/service_inter"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

const Name = "plugin"

type Middleware = service_inter.Middleware
type Plugin interface {
	json.Marshaler
	// String 插件描述
	String() string
	// ID 插件唯一名字
	ID() string
	// Flags 插件启动flags
	Flags() typex.Flags
	// Commands 插件启动子命令
	Commands() *typex.Command
	// Init 插件初始化
	Init(cfg config.Config) error
	// Watch 配置变更通知
	Watch(name string, r *watcher.Response) error
	// Vars 插件可观测指标
	Vars(vars.Publisher) error
	// Health 插件健康检查
	Health() healthy.Handler
	// Middleware 插件中间件拦截器
	Middleware() Middleware
	// BeforeStarts 在服务启动之前执行操作
	//	初始化, 检查, 注册, 上报等
	BeforeStarts() []func()
	// AfterStarts 在服务启动之后执行操作
	//	服务检查, 上报等
	AfterStarts() []func()
	// BeforeStops 在服务关闭之前执行操作
	//	关闭服务, 资源关闭等
	BeforeStops() []func()
	// AfterStops 在服务关闭之后执行操作
	//	关闭服务, 资源关闭等
	AfterStops() []func()
}

type Process interface {
	BeforeStart(fn func())
	AfterStart(fn func())
	BeforeStop(fn func())
	AfterStop(fn func())
}
