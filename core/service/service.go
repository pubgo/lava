// Package service 定义服务注册/发现领域模型（Service、Node、Instance 等）。
//
// 注意：本包的 Service 表示「注册中心里的一条服务记录」，
// 与 core/supervisor 中管理进程生命周期的 Service 接口不同。
package service

import (
	"strconv"
	"strings"
)

// Service 是注册中心中的一条服务记录，包含多个 Node。
type Service struct {
	Name      string            `json:"name,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Endpoints []string          `json:"endpoints,omitempty"`
	Nodes     []*Node           `json:"nodes,omitempty"`
}

type (
	Nodes []*Node
	// Node 表示一个可访问的服务实例节点。
	Node struct {
		Id       string            `json:"id,omitempty"`
		Version  string            `json:"version,omitempty"`
		Address  string            `json:"address,omitempty"`
		Port     int               `json:"port,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}
)

// GetPort 返回节点端口：优先使用 Port 字段，否则从 Address 的 host:port 后缀解析。
func (t Node) GetPort() int {
	if t.Port != 0 {
		return t.Port
	}

	parts := strings.Split(t.Address, ":")
	if len(parts) > 1 {
		port, _ := strconv.Atoi(parts[len(parts)-1])
		return port
	}

	return 0
}

// Endpoint 描述 RPC/HTTP 端点的请求/响应 schema。
type Endpoint struct {
	Name     string            `json:"name,omitempty"`
	Request  *Value            `json:"request,omitempty"`
	Response *Value            `json:"response,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Value 是 Endpoint schema 中的字段描述。
type Value struct {
	Name   string   `json:"name,omitempty"`
	Type   string   `json:"type,omitempty"`
	Values []*Value `json:"values,omitempty"`
}

// Instance 是 discovery 系统中一个服务实例的快照。
type Instance struct {
	// ID 是注册时分配的唯一实例 ID。
	ID string `json:"id"`
	// Name 是注册的服务名。
	Name string `json:"name"`
	// Version 是编译版本号。
	Version string `json:"version"`
	// Metadata 是实例关联的键值元数据。
	Metadata map[string]string `json:"metadata"`
	// Endpoints 是实例暴露的端点地址，例如：
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
}
