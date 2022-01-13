package plugin

import (
	"encoding/json"

	"github.com/pubgo/lava/types"
)

const Name = "plugin"

type Plugin interface {
	json.Marshaler
	// String 插件描述
	String() string
	// ID 插件唯一名字
	ID() string
	// Flags 插件启动flags
	Flags() types.Flags
	// Commands 插件启动子命令
	Commands() *types.Command
	// Init 插件初始化
	Init() error
	// Watch 配置变更通知
	Watch() types.Watcher
	// Vars 插件可观测指标
	Vars(types.Vars) error
	// Health 插件健康检查
	Health() types.Healthy
	// Middleware 插件中间件拦截器
	Middleware() types.Middleware
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
