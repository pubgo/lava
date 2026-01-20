package tunnel

// Config 代理网关配置
type Config struct {
	// Gateway 网关配置
	Gateway GatewayConfig `yaml:"gateway"`
	// Agent 代理客户端配置
	Agent AgentConfig `yaml:"agent"`
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled"`
	// ListenAddr 监听地址
	ListenAddr string `yaml:"listen_addr"`
	// Transport 传输协议: yamux, quic, http
	Transport string `yaml:"transport"`
	// TransportOptions 传输层选项
	TransportOptions *TransportOptions `yaml:"transport_options"`
	// HTTPPort HTTP 服务端口
	HTTPPort int `yaml:"http_port"`
	// GRPCPort gRPC 服务端口
	GRPCPort int `yaml:"grpc_port"`
	// DebugPort Debug 服务端口
	DebugPort int `yaml:"debug_port"`
	// HeartbeatInterval 心跳间隔（秒）
	HeartbeatInterval int `yaml:"heartbeat_interval"`
	// HeartbeatTimeout 心跳超时（秒）
	HeartbeatTimeout int `yaml:"heartbeat_timeout"`
	// HealthCheckInterval 健康检查间隔（秒）
	HealthCheckInterval int `yaml:"health_check_interval"`
	// TLS TLS 配置
	TLS TLSConfig `yaml:"tls"`
}

// AgentConfig 代理客户端配置
type AgentConfig struct {
	// Enabled 是否启用
	Enabled bool `yaml:"enabled"`
	// GatewayAddr 网关地址
	GatewayAddr string `yaml:"gateway_addr"`
	// Transport 传输协议
	Transport string `yaml:"transport"`
	// TransportOptions 传输层选项
	TransportOptions *TransportOptions `yaml:"transport_options"`
	// ServiceID 服务ID，如果为空则自动生成
	ServiceID string `yaml:"service_id"`
	// ServiceName 服务名称
	ServiceName string `yaml:"service_name"`
	// ServiceVersion 服务版本
	ServiceVersion string `yaml:"service_version"`
	// Metadata 服务元数据
	Metadata map[string]string `yaml:"metadata"`
	// Endpoints 要暴露的端点
	Endpoints []EndpointConfig `yaml:"endpoints"`
	// HeartbeatInterval 心跳间隔（秒）
	HeartbeatInterval int `yaml:"heartbeat_interval"`
	// ReconnectInterval 重连间隔（秒）
	ReconnectInterval int `yaml:"reconnect_interval"`
	// MaxReconnectAttempts 最大重连次数，0 表示无限重试
	MaxReconnectAttempts int `yaml:"max_reconnect_attempts"`
	// TLS TLS 配置
	TLS TLSConfig `yaml:"tls"`
}

// EndpointConfig 端点配置
type EndpointConfig struct {
	// Type 端点类型: http, grpc, debug
	Type string `yaml:"type"`
	// LocalAddr 本地地址
	LocalAddr string `yaml:"local_addr"`
	// Path 暴露路径
	Path string `yaml:"path"`
	// Metadata 端点元数据
	Metadata map[string]string `yaml:"metadata"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	// Enabled 是否启用 TLS
	Enabled bool `yaml:"enabled"`
	// CertFile 证书文件
	CertFile string `yaml:"cert_file"`
	// KeyFile 私钥文件
	KeyFile string `yaml:"key_file"`
	// CAFile CA 证书文件
	CAFile string `yaml:"ca_file"`
	// Insecure 是否跳过证书验证
	Insecure bool `yaml:"insecure"`
}

// TransportOptions 传输层选项
type TransportOptions struct {
	// EnableTLS 是否启用 TLS
	EnableTLS bool
	// CertFile 证书文件
	CertFile string
	// KeyFile 私钥文件
	KeyFile string
	// CAFile CA 证书文件
	CAFile string
	// Insecure 是否跳过证书验证
	Insecure bool
	// MaxStreams 最大流数量
	MaxStreams int
	// KeepAliveInterval 保活间隔
	KeepAliveInterval int
	// ConnectionWriteTimeout 连接写超时（秒）
	ConnectionWriteTimeout int
	// StreamOpenTimeout 流打开超时（秒）
	StreamOpenTimeout int
}

// DefaultTransportOptions 默认传输层选项
func DefaultTransportOptions() *TransportOptions {
	return &TransportOptions{
		MaxStreams:             256,
		KeepAliveInterval:      30,
		ConnectionWriteTimeout: 10,
		StreamOpenTimeout:      30,
	}
}

// DefaultGatewayConfig 默认网关配置
func DefaultGatewayConfig() GatewayConfig {
	return GatewayConfig{
		Enabled:             false,
		ListenAddr:          ":7007",
		Transport:           "yamux",
		HTTPPort:            8080,
		GRPCPort:            9090,
		DebugPort:           6060,
		HeartbeatInterval:   30,
		HeartbeatTimeout:    90,
		HealthCheckInterval: 30,
	}
}

// DefaultAgentConfig 默认代理客户端配置
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Enabled:              false,
		Transport:            "yamux",
		HeartbeatInterval:    30,
		ReconnectInterval:    5,
		MaxReconnectAttempts: 0, // 无限重试
	}
}
