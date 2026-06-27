package rpcmeta

import (
	"github.com/pubgo/funk/v2/assert"
)

// rpcMetas 同时以服务名（Name）和方法名（Method）为键索引 RpcMeta，
// 以便按任一标识快速查询。
var rpcMetas = make(map[string]*RpcMeta)

// Register 注册一个 RPC 方法的元信息。meta 不可为空，Name/Method 不可为空，
// 且 Name 与 Method 均不可重复注册，否则触发 panic。
func Register(meta *RpcMeta) error {
	assert.If(meta == nil, "rpc meta is nil")
	assert.If(meta.Name == "", "rpc meta name is empty")
	assert.If(meta.Method == "", "rpc meta method is nil")
	assert.If(rpcMetas[meta.Name] != nil, "rpc meta name already exists")
	assert.If(rpcMetas[meta.Method] != nil, "rpc meta method already exists")

	rpcMetas[meta.Name] = meta
	rpcMetas[meta.Method] = meta
	return nil
}

// Get 按服务名或方法名返回已注册的 RpcMeta，不存在时返回 nil。
func Get(nameOrMethod string) *RpcMeta { return rpcMetas[nameOrMethod] }
