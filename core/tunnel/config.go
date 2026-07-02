package tunnel

// GatewayConfig 网关配置
type GatewayConfig struct {
	// ListenAddr 监听地址
	ListenAddr string `yaml:"listen_addr"`
	// Transport 传输协议: yamux, quic, kcp
	Transport string `yaml:"transport"`
	// TransportOptions 传输层选项
	TransportOptions *TransportOptions `yaml:"transport_options"`
	// HTTPPort HTTP 服务端口
	HTTPPort int `yaml:"http_port"`
	// GRPCPort 是对外暴露的 gRPC 代理端口。
	// 客户端连接后须先发送路由行 "TUNNEL <service-name>\n"，再发送 gRPC/HTTP2 流量。
	GRPCPort int `yaml:"grpc_port"`
	// DebugPort Debug 服务端口
	DebugPort int `yaml:"debug_port"`
	// HeartbeatInterval 心跳间隔（秒）
	HeartbeatInterval int `yaml:"heartbeat_interval"`
	// HeartbeatTimeout 心跳超时（秒）
	HeartbeatTimeout int `yaml:"heartbeat_timeout"`
	// HealthCheckInterval 健康检查间隔（秒）
	HealthCheckInterval int `yaml:"health_check_interval"`
	// P2PSignalRateLimit 每 peer 每秒最大 P2P 信令条数（0 表示默认 60）
	P2PSignalRateLimit int `yaml:"p2p_signal_rate_limit"`
	// P2PRegisterRateLimit 每 agent 每秒最大 P2P 注册次数（0 表示默认 10）
	P2PRegisterRateLimit int `yaml:"p2p_register_rate_limit"`
	// TLS TLS 配置
	TLS TLSConfig `yaml:"tls"`
}

// AgentConfig 代理客户端配置
type AgentConfig struct {
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
	// P2PSignalHandler 收到 gateway 转发的 P2P 信令时回调（JSON 为 signaling.Message）。
	P2PSignalHandler func(payload []byte)
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
	// MinVersion 最小 TLS 版本 (e.g. "TLS12", "TLS13")
	MinVersion string `yaml:"min_version"`
	// CipherSuites 密码套件列表 (e.g. ["TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"])
	CipherSuites []string `yaml:"cipher_suites"`
	// ClientAuth 客户端认证模式 ("NoClientCert", "RequestClientCert", "RequireAnyClientCert", "VerifyClientCertIfGiven", "RequireAndVerifyClientCert")
	ClientAuth string `yaml:"client_auth"`
	// SessionCacheSize 会话缓存大小
	SessionCacheSize int `yaml:"session_cache_size"`
	// SessionTimeout 会话超时时间（秒）
	SessionTimeout int `yaml:"session_timeout"`
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
	// MinVersion 最小 TLS 版本 (TLS12, TLS13)
	MinVersion string
	// CipherSuites TLS 密码套件名列表
	CipherSuites []string
	// ClientAuth 服务端 mTLS 模式（见 TLSConfig.ClientAuth）
	ClientAuth string
	// SessionCacheSize TLS 会话缓存大小
	SessionCacheSize int
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
		Transport:            "yamux",
		HeartbeatInterval:    30,
		ReconnectInterval:    5,
		MaxReconnectAttempts: 0, // 无限重试
	}
}
