package rpcmeta

import (
	"github.com/pubgo/funk/assert"
)

var rpcMetas = make(map[string]*RpcMeta)

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

func Get(nameOrMethod string) *RpcMeta { return rpcMetas[nameOrMethod] }
