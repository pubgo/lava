// Package flags 提供 CLI 命令行参数（flag）的全局聚合注册表。
//
// 各模块在 init() 中通过 Register 注册自己的 flag，
// 命令入口（lavabuilder）再通过 GetFlags 统一收集并挂载到根命令上。
package flags

import (
	"github.com/pubgo/redant"
)

var flags []redant.Option

// Register 注册一个 CLI flag。通常在包的 init() 中调用。
func Register(flag redant.Option) {
	flags = append(flags, flag)
}

// GetFlags 返回所有已注册的 CLI flag。
func GetFlags() []redant.Option { return flags }
