// Package tunnel 提供服务代理网关的核心功能
//
// 这个包实现了一个服务注册监控网关系统，允许服务通过反向代理的方式注册到代理网关，
// 然后通过代理网关暴露服务的 API、gRPC、debug 调试等接口。
//
// 主要功能：
//   - 服务注册：服务启动后自动注册到代理网关
//   - 多协议传输：支持 yamux、QUIC、HTTP CONNECT tunnel、KCP 等
//   - 服务暴露：通过代理网关暴露 HTTP API、gRPC 和 debug 调试接口
//   - 健康检查：自动监控服务健康状态
//   - 运维监控：提供统一的调试、监控入口
//
// 架构设计：
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                      Proxy Gateway Server                        │
//	│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
//	│  │   HTTP API  │ │    gRPC     │ │    Debug    │               │
//	│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘               │
//	│         │               │               │                       │
//	│  ┌──────┴───────────────┴───────────────┴──────┐               │
//	│  │              Service Router                  │               │
//	│  └──────────────────────┬──────────────────────┘               │
//	│                         │                                       │
//	│  ┌──────────────────────┴──────────────────────┐               │
//	│  │             Transport Manager                │               │
//	│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐       │               │
//	│  │  │  yamux  │ │  QUIC   │ │  HTTP   │  ...  │               │
//	│  │  └─────────┘ └─────────┘ └─────────┘       │               │
//	│  └──────────────────────────────────────────────┘               │
//	└─────────────────────────────────────────────────────────────────┘
//	                           ↑
//	                     ┌─────┴─────┐
//	                     │  Tunnel   │
//	                     └─────┬─────┘
//	                           ↓
//	┌─────────────────────────────────────────────────────────────────┐
//	│                       Service Agent                              │
//	│  ┌──────────────────────┴──────────────────────┐               │
//	│  │             Transport Client                 │               │
//	│  └──────────────────────┬──────────────────────┘               │
//	│         ┌───────────────┼───────────────┐                       │
//	│  ┌──────┴──────┐ ┌──────┴──────┐ ┌──────┴──────┐               │
//	│  │   HTTP API  │ │    gRPC     │ │    Debug    │               │
//	│  └─────────────┘ └─────────────┘ └─────────────┘               │
//	└─────────────────────────────────────────────────────────────────┘
package tunnel

import (
	"context"
	"io"
	"net"
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	// ID 服务唯一标识
	ID string `json:"id,omitempty"`
	// Name 服务名称
	Name string `json:"name,omitempty"`
	// Version 服务版本
	Version string `json:"version,omitempty"`
	// Metadata 元数据
	Metadata map[string]string `json:"metadata,omitempty"`
	// Endpoints 服务端点列表
	Endpoints []Endpoint `json:"endpoints,omitempty"`
	// RegisterTime 注册时间
	RegisterTime time.Time `json:"register_time,omitempty"`
	// LastHeartbeat 最后心跳时间
	LastHeartbeat time.Time `json:"last_heartbeat,omitempty"`
	// Status 服务状态
	Status ServiceStatus `json:"status,omitempty"`
}

// Endpoint 服务端点
type Endpoint struct {
	// Type 端点类型: http, grpc, debug
	Type EndpointType `json:"type,omitempty"`
	// Path 端点路径
	Path string `json:"path,omitempty"`
	// Port 端点端口（本地）
	Port int `json:"port,omitempty"`
	// Address 端点地址
	Address string `json:"address,omitempty"`
	// Metadata 端点元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EndpointType 端点类型
type EndpointType string

const (
	EndpointTypeHTTP  EndpointType = "http"
	EndpointTypeGRPC  EndpointType = "grpc"
	EndpointTypeDebug EndpointType = "debug"
	EndpointTypeTCP   EndpointType = "tcp"
	EndpointTypeUDP   EndpointType = "udp"
)

// ServiceStatus 服务状态
type ServiceStatus string

const (
	ServiceStatusOnline    ServiceStatus = "online"
	ServiceStatusOffline   ServiceStatus = "offline"
	ServiceStatusUnhealthy ServiceStatus = "unhealthy"
)

// Transport 传输层接口，支持多种传输协议
// 实现者需要提供底层连接的多路复用能力
type Transport interface {
	// Name 传输协议名称
	Name() string
	// Dial 创建到服务端的连接
	Dial(ctx context.Context, addr string) (Session, error)
	// Listen 监听来自客户端的连接
	Listen(ctx context.Context, addr string) (Listener, error)
}

// Session 代表一个与对端的会话，可以创建多个流
type Session interface {
	io.Closer

	// Open 打开一个新的流
	Open(ctx context.Context) (Stream, error)
	// OpenWithPriority 打开指定优先级的流（1-10，1最高）
	OpenWithPriority(ctx context.Context, priority int) (Stream, error)
	// Accept 接受一个新的流
	Accept() (Stream, error)
	// IsClosed 会话是否已关闭
	IsClosed() bool
	// NumStreams 当前活跃流数量
	NumStreams() int
	// LocalAddr 本地地址
	LocalAddr() net.Addr
	// RemoteAddr 远程地址
	RemoteAddr() net.Addr
}

// Stream 代表一个多路复用的流连接
type Stream interface {
	io.ReadWriteCloser

	// LocalAddr 本地地址
	LocalAddr() net.Addr
	// RemoteAddr 远程地址
	RemoteAddr() net.Addr
	// SetDeadline 设置读写超时
	SetDeadline(t time.Time) error
	// SetReadDeadline 设置读超时
	SetReadDeadline(t time.Time) error
	// SetWriteDeadline 设置写超时
	SetWriteDeadline(t time.Time) error
	// Priority 获取流优先级
	Priority() int
}

// Listener 监听器接口
type Listener interface {
	io.Closer
	// Accept 接受新的会话连接
	Accept() (Session, error)
	// Addr 监听地址
	Addr() net.Addr
}

// Agent 服务代理客户端接口
// 运行在服务节点上，负责将本地服务注册到代理网关
type Agent interface {
	// Start 启动代理客户端
	Start(ctx context.Context) error
	// Stop 停止代理客户端
	Stop(ctx context.Context) error
	// Register 注册服务
	Register(ctx context.Context, service *ServiceInfo) error
	// Deregister 注销服务
	Deregister(ctx context.Context, serviceID string) error
	// Status 获取代理客户端状态
	Status() AgentStatus
	// Info 获取 Agent 信息（用于调试显示）
	Info() *AgentInfo
}

// AgentInfo Agent 信息
type AgentInfo struct {
	GatewayAddr    string     `json:"gateway_addr"`
	ServiceName    string     `json:"service_name"`
	ServiceVersion string     `json:"service_version"`
	Endpoints      []Endpoint `json:"endpoints"`
	Status         string     `json:"status"`
}

// AgentStatus 代理客户端状态
type AgentStatus int

const (
	// StatusDisconnected 未连接
	StatusDisconnected AgentStatus = iota
	// StatusConnecting 连接中
	StatusConnecting
	// StatusConnected 已连接
	StatusConnected
	// StatusReconnecting 重连中
	StatusReconnecting
)

// String 返回状态的字符串表示
func (s AgentStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// AgentStatusInfo 代理客户端状态信息
type AgentStatusInfo struct {
	// Connected 是否已连接到网关
	Connected bool `json:"connected"`
	// GatewayAddr 网关地址
	GatewayAddr string `json:"gateway_addr"`
	// Transport 使用的传输协议
	Transport string `json:"transport"`
	// Services 已注册的服务列表
	Services []ServiceInfo `json:"services"`
	// LastError 最后一次错误
	LastError string `json:"last_error,omitempty"`
	// ConnectedAt 连接时间
	ConnectedAt time.Time `json:"connected_at,omitempty"`
}

// AuthProvider 认证提供者接口
type AuthProvider interface {
	// Authenticate 验证服务是否可以注册
	Authenticate(service *ServiceInfo) error
	// Authorize 验证客户端是否可以访问服务
	Authorize(serviceID, clientID string) error
	// GenerateToken 生成认证令牌
	GenerateToken(service *ServiceInfo) (string, error)
	// ValidateToken 验证认证令牌
	ValidateToken(token string) (*ServiceInfo, error)
}

// Gateway 代理网关服务端接口
// 运行在代理网关上，负责接收服务注册并暴露服务
type Gateway interface {
	// Start 启动网关
	Start(ctx context.Context) error
	// Stop 停止网关
	Stop(ctx context.Context) error
	// Services 获取所有注册的服务
	Services() []*ServiceInfo
	// GetService 获取指定服务
	GetService(name string) (*ServiceInfo, error)
	// Status 获取网关状态
	Status() GatewayStatus
	// Forward 转发请求到指定服务
	Forward(ctx context.Context, serviceName string, endpointType EndpointType, conn net.Conn) error
	// SetAuthProvider 设置认证提供者
	SetAuthProvider(auth AuthProvider)
}

// GatewayStatus 网关状态
type GatewayStatus int

const (
	// GatewayStatusStopped 网关已停止
	GatewayStatusStopped GatewayStatus = iota
	// GatewayStatusStarting 网关启动中
	GatewayStatusStarting
	// GatewayStatusRunning 网关运行中
	GatewayStatusRunning
	// GatewayStatusStopping 网关停止中
	GatewayStatusStopping
)

// String 返回状态的字符串表示
func (s GatewayStatus) String() string {
	switch s {
	case GatewayStatusStopped:
		return "stopped"
	case GatewayStatusStarting:
		return "starting"
	case GatewayStatusRunning:
		return "running"
	case GatewayStatusStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// GatewayStatusInfo 网关状态信息
type GatewayStatusInfo struct {
	// Running 是否运行中
	Running bool `json:"running"`
	// ListenAddr 监听地址
	ListenAddr string `json:"listen_addr"`
	// Transport 使用的传输协议
	Transport string `json:"transport"`
	// ServiceCount 注册的服务数量
	ServiceCount int `json:"service_count"`
	// ConnectionCount 当前连接数
	ConnectionCount int `json:"connection_count"`
	// StartedAt 启动时间
	StartedAt time.Time `json:"started_at,omitempty"`
}

// Message 通信消息（类似 JSON-RPC）
type Message struct {
	// Type 消息类型/方法
	Type MessageType `json:"type"`
	// ID 消息ID，用于请求-响应匹配
	ID string `json:"id,omitempty"`
	// Payload 消息负载（JSON 编码的具体数据）
	Payload []byte `json:"payload,omitempty"`
	// Error 错误信息
	Error string `json:"error,omitempty"`
}

// MessageType 消息类型
type MessageType uint8

const (
	// MessageTypeRegister 服务注册
	MessageTypeRegister MessageType = iota + 1
	// MessageTypeDeregister 服务注销
	MessageTypeDeregister
	// MessageTypeHeartbeat 心跳
	MessageTypeHeartbeat
	// MessageTypeRequest 请求
	MessageTypeRequest
	// MessageTypeResponse 响应
	MessageTypeResponse
	// MessageTypeStream 流数据
	MessageTypeStream
	// MessageTypeAck 确认
	MessageTypeAck
	// MessageTypeError 错误
	MessageTypeError
	// MessageTypeHTTPRequest HTTP请求转发
	MessageTypeHTTPRequest
	// MessageTypeGRPCRequest gRPC请求转发
	MessageTypeGRPCRequest
	// MessageTypeDebugRequest Debug请求转发
	MessageTypeDebugRequest
)

// RequestMeta 请求元数据
type RequestMeta struct {
	// ServiceID 目标服务ID
	ServiceID string `json:"service_id"`
	// EndpointType 端点类型
	EndpointType EndpointType `json:"endpoint_type"`
	// Path 请求路径
	Path string `json:"path"`
	// Method HTTP方法（仅HTTP端点）
	Method string `json:"method,omitempty"`
	// Headers 请求头
	Headers map[string]string `json:"headers,omitempty"`
}
