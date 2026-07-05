package tunnelgateway

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

var _ tunnel.Gateway = (*tunnelGateway)(nil)

// NewGateway creates a new tunnel gateway
func NewGateway(cfg *tunnel.GatewayConfig) tunnel.Gateway {
	if cfg != nil {
		cfg.Normalize()
	}
	return &tunnelGateway{
		cfg:                cfg,
		services:           make(map[string]*registeredService),
		peers:              make(map[string]*registeredPeer),
		peerByAgent:        make(map[string]string),
		status:             tunnel.GatewayStatusStopped,
		rateLimiter:        ratelimit.New(100),
		p2pSignalLimiter:   ratelimit.New(p2pSignalRateLimit(cfg)),
		p2pRegisterLimiter: ratelimit.New(p2pRegisterRateLimit(cfg)),
	}
}

func p2pSignalRateLimit(cfg *tunnel.GatewayConfig) int {
	const defaultLimit = 60
	if cfg == nil || cfg.P2PSignalRateLimit <= 0 {
		return defaultLimit
	}
	return cfg.P2PSignalRateLimit
}

func p2pRegisterRateLimit(cfg *tunnel.GatewayConfig) int {
	const defaultLimit = 10
	if cfg == nil || cfg.P2PRegisterRateLimit <= 0 {
		return defaultLimit
	}
	return cfg.P2PRegisterRateLimit
}

type registeredService struct {
	info    *tunnel.ServiceInfo
	session tunnel.Session
	agent   string // agent identifier
}

type registeredPeer struct {
	peerID       string
	session      tunnel.Session
	agentID      string
	registeredAt time.Time
}

type tunnelGateway struct {
	cfg                *tunnel.GatewayConfig
	transport          tunnel.Transport
	listener           tunnel.Listener
	services           map[string]*registeredService
	peers              map[string]*registeredPeer
	peerByAgent        map[string]string
	status             tunnel.GatewayStatus
	authProvider       tunnel.AuthProvider
	metrics            *tunnel.MetricsRecorder
	rateLimiter        *ratelimit.Limiter
	p2pSignalLimiter   *ratelimit.Limiter
	p2pRegisterLimiter *ratelimit.Limiter

	// 对外代理服务器
	httpServer   *http.Server
	debugServer  *http.Server
	grpcListener net.Listener

	mu       sync.RWMutex
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	running  atomic.Bool
}

// SetAuthProvider 设置认证提供者
func (g *tunnelGateway) SetAuthProvider(auth tunnel.AuthProvider) {
	g.authProvider = auth
}

// SetMetricsRecorder 绑定可选 metrics 记录器。
func (g *tunnelGateway) SetMetricsRecorder(m *tunnel.MetricsRecorder) {
	g.metrics = m
}
