// Package healthy 提供健康检查处理器（Handler）的全局注册表。
//
// 各组件可在初始化时通过 Register 注册自己的健康检查逻辑，
// 健康检查端点再通过 List/Get 汇总并执行这些 Handler。
package healthy

import (
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/typex"
)

const Name = "healthy"

var healthList typex.SyncMap

// Get 返回指定名称的健康检查 Handler，不存在时返回 nil。
func Get(name string) Handler {
	val, ok := healthList.Load(name)
	if !ok {
		return nil
	}

	return val.(Handler)
}

// List 返回所有已注册的健康检查名称。
func List() (names []string) {
	healthList.Range(func(name, _ any) bool {
		names = append(names, name.(string))
		return true
	})
	return names
}

// Register 注册一个健康检查 Handler。name 与 r 均不可为空，
// 且 name 不可重复注册，否则触发 panic。
func Register(name string, r Handler) {
	assert.If(name == "" || r == nil, "[name,r] is null")
	assert.If(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
