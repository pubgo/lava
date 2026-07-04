// Package rpcmeta 定义描述单个 RPC 方法的元信息结构。
//
// 它在代码生成与服务注册阶段用于携带方法名、所属服务名、
// 自定义标签以及请求/响应的 proto 消息原型。
package rpcmeta

import "google.golang.org/protobuf/proto"

// RpcMeta 描述一个 RPC 方法的元信息。
type RpcMeta struct {
	// Method 是方法名。
	Method string
	// Name 是所属服务名。
	Name string
	// Tags 是附加在该方法上的自定义键值标签。
	Tags map[string]string
	// Input 是请求消息的 proto 原型。
	Input proto.Message
	// Output 是响应消息的 proto 原型。
	Output proto.Message
}
