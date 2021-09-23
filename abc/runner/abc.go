package runner

import "context"

const Name = "runner"

/*
1. Runner定义了一个可运行的模块, 实现了Runner的模块会被编译成plugin, 然后加载到项目中
*/

// Runner 可运行的模块
type Runner interface {
	Init() error
	Name() string
	Register(name string, fn interface{}) error
	Run(ctx context.Context, req interface{}, names ...string) (resp interface{}, err error)
}
