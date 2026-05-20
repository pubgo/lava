package tunnel

import "errors"

var (
	// ErrSessionClosed 会话已关闭
	ErrSessionClosed = errors.New("tunnel: session closed")
	// ErrStreamClosed 流已关闭
	ErrStreamClosed = errors.New("tunnel: stream closed")
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("tunnel: connection failed")
	// ErrServiceNotFound 服务未找到
	ErrServiceNotFound = errors.New("tunnel: service not found")
	// ErrServiceAlreadyExists 服务已存在
	ErrServiceAlreadyExists = errors.New("tunnel: service already exists")
	// ErrInvalidMessage 无效消息
	ErrInvalidMessage = errors.New("tunnel: invalid message")
	// ErrTimeout 超时
	ErrTimeout = errors.New("tunnel: timeout")
	// ErrTransportNotSupported 不支持的传输协议
	ErrTransportNotSupported = errors.New("tunnel: transport not supported")
	// ErrGatewayNotConnected 未连接到网关
	ErrGatewayNotConnected = errors.New("tunnel: gateway not connected")
	// ErrAgentNotRunning 代理客户端未运行
	ErrAgentNotRunning = errors.New("tunnel: agent not running")
	// ErrAgentAlreadyRunning 代理客户端已在运行
	ErrAgentAlreadyRunning = errors.New("tunnel: agent already running")
	// ErrGatewayAlreadyRunning 网关已在运行
	ErrGatewayAlreadyRunning = errors.New("tunnel: gateway already running")
)
