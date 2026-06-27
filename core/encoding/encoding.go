// Package encoding 提供编解码器（Codec）的全局注册表，
// 用于按名称（如 "json"、"proto"）或 HTTP Content-Type 选择对应的序列化实现。
//
// 各编解码器实现通常在自身包的 init() 中调用 Register 完成注册，
// 调用方通过 Get / GetWithCT 获取对应 Codec。
package encoding

import (
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/typex"
)

var data typex.Map[Codec]

// Register 注册一个编解码器。name 与 cdc 均不可为空，且 name 不可重复注册，
// 否则会触发 panic（通过 recovery.Exit 统一处理）。
func Register(name string, cdc Codec) {
	defer recovery.Exit()
	assert.If(cdc == nil || name == "" || cdc.Name() == "", "codec[%s] is null", name)
	assert.If(data.Has(name), "[cdc] %s already exists", name)
	data.Set(name, cdc)
}

// Get 按名称返回已注册的编解码器，不存在时返回 nil。
func Get(name string) Codec {
	val, ok := data.Load(name)
	if !ok {
		return nil
	}

	return val
}

// Keys 返回所有已注册编解码器的名称。
func Keys() []string { return data.Keys() }

// Each 遍历所有已注册的编解码器。
func Each(fn func(name string, cdc Codec)) {
	defer recovery.Exit()

	data.Each(func(name string, val Codec) {
		fn(name, val)
	})
}

// GetWithCT 按 HTTP Content-Type 返回对应的编解码器，无匹配时返回 nil。
func GetWithCT(ct string) Codec {
	return Get(cdcMapping[ct])
}

// cdcMapping 维护 HTTP Content-Type 到编解码器名称的映射。
var cdcMapping = map[string]string{
	"application/json":         "json",
	"application/proto":        "proto",
	"application/protobuf":     "proto",
	"application/octet-stream": "proto",
	"application/grpc":         "proto",
	"application/grpc+json":    "json",
	"application/grpc+proto":   "proto",
	"application/grpc+bytes":   "bytes",
}
