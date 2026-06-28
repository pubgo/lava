package tunnelgateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

// linkRewritePatterns 用于重写 HTML 响应中的链接
var linkRewritePatterns = []*regexp.Regexp{
	// href="/debug/..." or href='/debug/...'
	regexp.MustCompile(`(href=["'])/debug/`),
	// src="/debug/..." or src='/debug/...'
	regexp.MustCompile(`(src=["'])/debug/`),
	// action="/debug/..."
	regexp.MustCompile(`(action=["'])/debug/`),
	// fetch("/debug/..." or fetch('/debug/...'
	regexp.MustCompile(`(fetch\(["'])/debug/`),
	// url: "/debug/..." (for JavaScript)
	regexp.MustCompile(`(url:\s*["'])/debug/`),
	// "/debug/ in JSON responses
	regexp.MustCompile(`(")/debug/`),
}

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
	cfg          *tunnel.GatewayConfig
	transport    tunnel.Transport
	listener     tunnel.Listener
	services     map[string]*registeredService
	peers        map[string]*registeredPeer
	peerByAgent  map[string]string
	status       tunnel.GatewayStatus
	authProvider tunnel.AuthProvider
	metrics      *tunnel.MetricsRecorder
	rateLimiter  *ratelimit.Limiter
	p2pSignalLimiter   *ratelimit.Limiter
	p2pRegisterLimiter *ratelimit.Limiter

	// 对外代理服务器
	httpServer  *http.Server
	debugServer *http.Server
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

func (g *tunnelGateway) Start(ctx context.Context) error {
	if g.running.Load() {
		return tunnel.ErrGatewayAlreadyRunning
	}

	// Create transport
	transport, err := tunnel.NewTransport(g.cfg.Transport, g.cfg.TransportOptions)
	if err != nil {
		return err
	}
	g.transport = transport

	// Start listening for agent connections (被动接受 Agent 连接)
	listener, err := transport.Listen(ctx, g.cfg.ListenAddr)
	if err != nil {
		return err
	}
	g.listener = listener

	g.stopCh = make(chan struct{})
	g.running.Store(true)
	g.status = tunnel.GatewayStatusRunning

	// Start accepting agent connections
	g.wg.Add(1)
	go g.acceptLoop()

	// Start health check loop
	g.wg.Add(1)
	go g.healthCheckLoop()

	// Start HTTP proxy server (对外暴露 HTTP 端口)
	if g.cfg.HTTPPort > 0 {
		g.wg.Add(1)
		go g.startHTTPProxy()
	}

	// Start Debug proxy server (对外暴露 Debug 端口)
	if g.cfg.DebugPort > 0 {
		g.wg.Add(1)
		go g.startDebugProxy()
	}

	// Start gRPC proxy server (对外暴露 gRPC 端口)
	if g.cfg.GRPCPort > 0 {
		g.wg.Add(1)
		go g.startGRPCProxy()
	}

	log.Info().
		Str("tunnel_addr", g.cfg.ListenAddr).
		Int("http_port", g.cfg.HTTPPort).
		Int("grpc_port", g.cfg.GRPCPort).
		Int("debug_port", g.cfg.DebugPort).
		Msg("Gateway started, waiting for agents to connect...")

	return nil
}

// startHTTPProxy 启动 HTTP 代理服务器，接收外部 HTTP 请求并转发到 Agent
func (g *tunnelGateway) startHTTPProxy() {
	defer g.wg.Done()

	addr := fmt.Sprintf(":%d", g.cfg.HTTPPort)
	g.httpServer = &http.Server{
		Addr:    addr,
		Handler: g.createProxyHandler(tunnel.EndpointTypeHTTP),
	}

	log.Info().Str("addr", addr).Msg("HTTP proxy server started")

	if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("HTTP proxy server error")
	}
}

// startDebugProxy 启动 Debug 代理服务器，接收外部 Debug 请求并转发到 Agent
func (g *tunnelGateway) startDebugProxy() {
	defer g.wg.Done()

	addr := fmt.Sprintf(":%d", g.cfg.DebugPort)
	g.debugServer = &http.Server{
		Addr:    addr,
		Handler: g.createProxyHandler(tunnel.EndpointTypeDebug),
	}

	log.Info().Str("addr", addr).Msg("Debug proxy server started")

	if err := g.debugServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Debug proxy server error")
	}
}

// createProxyHandler 创建 HTTP 代理处理器
// URL 格式: /{service_name}/path... -> 转发到对应 Agent 的本地服务
func (g *tunnelGateway) createProxyHandler(endpointType tunnel.EndpointType) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析 URL: /{service_name}/path...
		path := r.URL.Path
		if !strings.HasPrefix(path, "/") {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		switch strings.TrimSuffix(path, "/") {
		case "/p2p/peers":
			if err := g.authorizeClient("", tunnel.ClientTokenFromRequest(r)); err != nil {
				writeAuthError(w)
				return
			}
			g.handlePeerList(w, r)
			return
		}

		parts := strings.SplitN(path[1:], "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			if err := g.authorizeClient("", tunnel.ClientTokenFromRequest(r)); err != nil {
				writeAuthError(w)
				return
			}
			g.handleServiceList(w, r)
			return
		}

		serviceName := parts[0]
		subPath := "/"
		if len(parts) > 1 {
			subPath = "/" + parts[1]
		}

		// 查找服务
		g.mu.RLock()
		svc, ok := g.services[serviceName]
		g.mu.RUnlock()

		if !ok {
			http.Error(w, fmt.Sprintf("service not found: %s", serviceName), http.StatusNotFound)
			return
		}

		if svc.session == nil || svc.session.IsClosed() {
			http.Error(w, fmt.Sprintf("service unavailable: %s", serviceName), http.StatusServiceUnavailable)
			return
		}

		if err := g.authorizeClient(serviceName, tunnel.ClientTokenFromRequest(r)); err != nil {
			writeAuthError(w)
			return
		}

		// 转发请求到 Agent
		g.proxyToAgent(w, r, svc, endpointType, subPath)
	})
}

// handleServiceList 返回已注册的服务列表
func (g *tunnelGateway) handleServiceList(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	services := make([]*tunnel.ServiceInfo, 0, len(g.services))
	for _, svc := range g.services {
		services = append(services, svc.info)
	}
	g.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"services": services,
		"count":    len(services),
	}); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to encode service list")
	}
}

type peerListEntry struct {
	PeerID       string    `json:"peer_id"`
	AgentID      string    `json:"agent_id"`
	RegisteredAt time.Time `json:"registered_at"`
}

// handlePeerList 返回已注册 P2P peer 列表（gateway 侧 debug）。
func (g *tunnelGateway) handlePeerList(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	peers := make([]peerListEntry, 0, len(g.peers))
	for _, p := range g.peers {
		peers = append(peers, peerListEntry{
			PeerID:       p.peerID,
			AgentID:      p.agentID,
			RegisteredAt: p.registeredAt,
		})
	}
	g.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"peers": peers,
		"count": len(peers),
	}); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to encode peer list")
	}
}

// proxyToAgent 将 HTTP 请求代理到 Agent
func (g *tunnelGateway) proxyToAgent(w http.ResponseWriter, r *http.Request, svc *registeredService, endpointType tunnel.EndpointType, subPath string) {
	ctx := r.Context()
	serviceName := svc.info.Name

	// 检查是否为 WebSocket 升级请求
	isWebSocket := strings.EqualFold(r.Header.Get("Upgrade"), "websocket")

	// 检查速率限制
	if !g.rateLimiter.Allow(serviceName) {
		log.Warn().Str("service", serviceName).Msg("Rate limit exceeded")
		if g.metrics != nil {
			g.metrics.ObserveRateLimited(serviceName)
		}
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if g.metrics != nil {
		g.metrics.ObserveProxyRequest(serviceName, string(endpointType))
	}

	log.Debug().
		Str("service", serviceName).
		Str("endpointType", string(endpointType)).
		Str("subPath", subPath).
		Str("method", r.Method).
		Bool("isWebSocket", isWebSocket).
		Bool("sessionClosed", svc.session.IsClosed()).
		Int("numStreams", svc.session.NumStreams()).
		Msg("Gateway: Proxying request to agent")

	// 确定流优先级
	priority := 5 // 默认中优先级
	switch endpointType {
	case tunnel.EndpointTypeDebug:
		priority = 3 // 调试请求使用较高优先级
	case tunnel.EndpointTypeGRPC:
		priority = 4 // gRPC 请求使用中等优先级
	}

	// 打开到 Agent 的 stream
	stream, err := svc.session.OpenWithPriority(ctx, priority)
	if err != nil {
		// 降级到普通优先级
		stream, err = svc.session.Open(ctx)
		if err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: Failed to open stream to agent")
			http.Error(w, fmt.Sprintf("failed to open stream: %v", err), http.StatusInternalServerError)
			return
		}
	}

	log.Debug().Str("service", serviceName).Int("priority", priority).Msg("Gateway: Stream opened to agent")

	// Build request meta
	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: endpointType,
		Path:         subPath,
		Method:       r.Method,
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to marshal request meta")
		http.Error(w, fmt.Sprintf("failed to build request meta: %v", err), http.StatusInternalServerError)
		return
	}

	// 发送请求消息给 Agent
	msg := &tunnel.Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Payload: payload,
	}

	if err := g.sendMessage(stream, msg); err != nil {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
		log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: Failed to send message to agent")
		http.Error(w, fmt.Sprintf("failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debug().Str("service", serviceName).Str("msgType", string(msg.Type)).Msg("Gateway: Message sent to agent, starting proxy")

	// 修改请求路径为子路径
	r.URL.Path = subPath
	r.RequestURI = subPath
	if r.URL.RawQuery != "" {
		r.RequestURI = subPath + "?" + r.URL.RawQuery
	}

	// WebSocket 请求使用双向 TCP 代理
	if isWebSocket {
		g.proxyWebSocket(w, r, stream, serviceName)
		return
	}

	// 普通 HTTP 请求使用 httputil 代理
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// 保持原始请求
		},
		Transport: &streamRoundTripper{stream: stream, request: r},
		ModifyResponse: func(resp *http.Response) error {
			// 只对 HTML 和 JSON 响应重写链接
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/json") {
				return nil
			}

			// 读取响应体
			body, err := io.ReadAll(resp.Body)
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Str("service", serviceName).Msg("Gateway: failed to close response body")
			}
			if err != nil {
				return err
			}

			// 重写链接: /debug/ -> /{serviceName}/debug/
			replacement := "${1}/" + serviceName + "/debug/"
			for _, pattern := range linkRewritePatterns {
				body = pattern.ReplaceAll(body, []byte(replacement))
			}

			// 更新响应体和 Content-Length
			resp.Body = io.NopCloser(bytes.NewReader(body))
			resp.ContentLength = int64(len(body))
			resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
			// 移除 Content-Encoding，因为我们已经解压了
			resp.Header.Del("Content-Encoding")

			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Warn().Err(err).Str("service", serviceName).Msg("Proxy error")
			w.WriteHeader(http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}

// proxyWebSocket 处理 WebSocket 代理
func (g *tunnelGateway) proxyWebSocket(w http.ResponseWriter, r *http.Request, stream tunnel.Stream, serviceName string) {
	log.Debug().Str("service", serviceName).Str("path", r.URL.Path).Msg("Gateway: Starting WebSocket proxy")

	// 获取底层 TCP 连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
		log.Warn().Err(err).Str("service", serviceName).Msg("Failed to hijack connection")
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}

	// 将原始 HTTP 请求写入 stream，让 Agent 处理 WebSocket 升级
	if err := r.Write(stream); err != nil {
		if err := clientConn.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close client connection")
		}
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
		log.Warn().Err(err).Str("service", serviceName).Msg("Failed to write WebSocket request to stream")
		return
	}

	log.Debug().Str("service", serviceName).Msg("Gateway: WebSocket request forwarded, starting bidirectional copy")

	// 双向复制数据
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Agent
	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, clientConn); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: websocket copy client->agent failed")
		}
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

	// Agent -> Client
	go func() {
		defer wg.Done()
		if _, err := io.Copy(clientConn, stream); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: websocket copy agent->client failed")
		}
		if err := clientConn.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close client connection")
		}
	}()

	wg.Wait()
	log.Debug().Str("service", serviceName).Msg("Gateway: WebSocket proxy finished")
}

// streamRoundTripper 实现 http.RoundTripper，通过 stream 转发请求
type streamRoundTripper struct {
	stream  tunnel.Stream
	request *http.Request
}

func (t *streamRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 将请求写入 stream
	if err := req.Write(t.stream); err != nil {
		return nil, err
	}

	// 从 stream 读取响应
	return http.ReadResponse(bufio.NewReader(t.stream), req)
}

func (g *tunnelGateway) Stop(ctx context.Context) error {
	if !g.running.Load() {
		return nil
	}

	g.stopOnce.Do(func() {
		close(g.stopCh)
	})

	// 关闭 HTTP 代理服务器
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if g.httpServer != nil {
		if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to shutdown HTTP server")
		}
	}
	if g.debugServer != nil {
		if err := g.debugServer.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to shutdown debug server")
		}
	}
	if g.grpcListener != nil {
		if err := g.grpcListener.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close GRPC listener")
		}
	}

	// Close listener
	if g.listener != nil {
		if err := g.listener.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close listener")
		}
	}

	// Close all sessions
	g.mu.Lock()
	for _, svc := range g.services {
		if svc.session != nil {
			if err := svc.session.Close(); err != nil {
				log.Warn().Err(err).Str("service", svc.info.Name).Msg("Gateway: failed to close session")
			}
		}
	}
	g.services = make(map[string]*registeredService)
	g.mu.Unlock()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		log.Warn().Msg("Gateway stop timeout, force closing")
	}

	g.running.Store(false)
	g.status = tunnel.GatewayStatusStopped
	return nil
}

func (g *tunnelGateway) Services() []*tunnel.ServiceInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	services := make([]*tunnel.ServiceInfo, 0, len(g.services))
	for _, svc := range g.services {
		services = append(services, svc.info)
	}
	return services
}

func (g *tunnelGateway) GetService(name string) (*tunnel.ServiceInfo, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	svc, ok := g.services[name]
	if !ok {
		return nil, tunnel.ErrServiceNotFound
	}
	return svc.info, nil
}

func (g *tunnelGateway) Status() tunnel.GatewayStatus {
	return g.status
}

func (g *tunnelGateway) Forward(ctx context.Context, serviceName string, endpointType tunnel.EndpointType, conn net.Conn) error {
	g.mu.RLock()
	svc, ok := g.services[serviceName]
	g.mu.RUnlock()

	if !ok {
		return tunnel.ErrServiceNotFound
	}

	if svc.session == nil || svc.session.IsClosed() {
		return tunnel.ErrSessionClosed
	}

	// Open a stream to the agent
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

	// Build request meta
	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: endpointType,
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// Send forward request
	msg := &tunnel.Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Payload: payload,
	}

	if err := g.sendMessage(stream, msg); err != nil {
		return err
	}

	// Bidirectional copy
	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(stream, conn)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(conn, stream)
		errCh <- err
	}()

	// Wait for one direction to complete
	<-errCh
	return nil
}

func (g *tunnelGateway) endpointTypeToMessageType(et tunnel.EndpointType) tunnel.MessageType {
	switch et {
	case tunnel.EndpointTypeHTTP:
		return tunnel.MessageTypeHTTPRequest
	case tunnel.EndpointTypeGRPC:
		return tunnel.MessageTypeGRPCRequest
	case tunnel.EndpointTypeDebug:
		return tunnel.MessageTypeDebugRequest
	default:
		return tunnel.MessageTypeHTTPRequest
	}
}

func (g *tunnelGateway) sendMessage(stream tunnel.Stream, msg *tunnel.Message) error {
	return tunnel.WriteMessage(stream, msg)
}

func (g *tunnelGateway) acceptLoop() {
	defer g.wg.Done()

	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		session, err := g.listener.Accept()
		if err != nil {
			select {
			case <-g.stopCh:
				return
			default:
				log.Warn().Err(err).Msg("Failed to accept connection")
				continue
			}
		}

		go g.handleSession(session)
	}
}

func (g *tunnelGateway) handleSession(session tunnel.Session) {
	agentID := session.RemoteAddr().String()
	log.Info().Str("agent", agentID).Msg("Agent connected")

	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		if session.IsClosed() {
			g.removeAgentServices(agentID)
			return
		}

		stream, err := session.Accept()
		if err != nil {
			if session.IsClosed() {
				g.removeAgentServices(agentID)
				return
			}
			log.Warn().Err(err).Msg("Failed to accept stream")
			continue
		}

		go g.handleStream(agentID, session, stream)
	}
}

func (g *tunnelGateway) handleStream(agentID string, session tunnel.Session, stream tunnel.Stream) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Msg("Gateway: failed to close stream")
		}
	}()

	// Read message (length-prefixed JSON, size-limited to avoid OOM)
	msg, err := tunnel.ReadMessage(stream)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read message")
		return
	}

	switch msg.Type {
	case tunnel.MessageTypeRegister:
		g.handleRegister(agentID, session, msg)
	case tunnel.MessageTypeDeregister:
		g.handleDeregister(agentID, msg)
	case tunnel.MessageTypeHeartbeat:
		g.handleHeartbeat(agentID)
	case tunnel.MessageTypeP2PRegister:
		g.handleP2PRegister(agentID, session, msg)
	case tunnel.MessageTypeP2PSignal:
		g.handleP2PSignal(agentID, session, msg)
	default:
		log.Warn().Uint8("type", uint8(msg.Type)).Msg("Unknown message type")
	}
}

func (g *tunnelGateway) handleRegister(agentID string, session tunnel.Session, msg *tunnel.Message) {
	if len(msg.Payload) == 0 {
		log.Warn().Str("agent", agentID).Msg("Register: empty payload")
		return
	}

	var service tunnel.ServiceInfo
	if err := json.Unmarshal(msg.Payload, &service); err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("Register: failed to parse service info")
		return
	}

	// 使用认证提供者验证服务
	if g.authProvider != nil {
		if err := g.authProvider.Authenticate(&service); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Str("service", service.Name).Msg("Register: authentication failed")
			return
		}
	}

	now := time.Now()
	service.RegisterTime = now
	service.LastHeartbeat = now
	service.Status = tunnel.ServiceStatusOnline

	g.mu.Lock()
	g.services[service.Name] = &registeredService{
		info:    &service,
		session: session,
		agent:   agentID,
	}
	registered := len(g.services)
	g.mu.Unlock()

	if g.metrics != nil {
		g.metrics.ObserveRegister(service.Name)
		g.metrics.SetRegisteredServices(registered)
	}

	log.Info().Str("service", service.Name).Str("agent", agentID).Msg("Service registered")
}

func (g *tunnelGateway) handleDeregister(agentID string, msg *tunnel.Message) {
	if len(msg.Payload) == 0 {
		return
	}

	serviceName := string(msg.Payload)

	g.mu.Lock()
	defer g.mu.Unlock()

	svc, ok := g.services[serviceName]
	if !ok {
		return
	}
	if svc.agent != agentID {
		log.Warn().
			Str("service", serviceName).
			Str("agent", agentID).
			Str("owner", svc.agent).
			Msg("Deregister: rejected (agent mismatch)")
		return
	}
	delete(g.services, serviceName)

	log.Info().Str("service", serviceName).Str("agent", agentID).Msg("Service deregistered")
}

func (g *tunnelGateway) handleHeartbeat(agentID string) {
	// 更新该 agent 名下所有服务的最后心跳时间，用于心跳超时检测
	now := time.Now()
	g.mu.Lock()
	for _, svc := range g.services {
		if svc.agent == agentID {
			svc.info.LastHeartbeat = now
		}
	}
	g.mu.Unlock()

	log.Debug().Str("agent", agentID).Msg("Heartbeat received")
}

func (g *tunnelGateway) removeAgentServices(agentID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for name, svc := range g.services {
		if svc.agent == agentID {
			delete(g.services, name)
			log.Info().Str("service", name).Str("agent", agentID).Msg("Service removed (agent disconnected)")
		}
	}
	g.removePeerLocked(agentID)
}

func (g *tunnelGateway) handleP2PRegister(agentID string, session tunnel.Session, msg *tunnel.Message) {
	if !g.p2pRegisterLimiter.Allow(agentID) {
		log.Warn().Str("agent", agentID).Msg("P2P register: rate limit exceeded")
		return
	}
	payload, err := tunnel.DecodeP2PRegisterPayload(msg.Payload)
	if err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("P2P register: invalid payload")
		return
	}
	if payload.PeerID == "" {
		log.Warn().Str("agent", agentID).Msg("P2P register: empty peer_id")
		return
	}
	if g.authProvider != nil {
		if payload.AuthToken == "" {
			log.Warn().Str("peer", payload.PeerID).Msg("P2P register: missing auth_token")
			return
		}
		if _, err := g.authProvider.ValidateToken(payload.AuthToken); err != nil {
			log.Warn().Err(err).Str("peer", payload.PeerID).Msg("P2P register: auth failed")
			return
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if oldAgent, ok := g.peerByAgent[agentID]; ok && oldAgent != payload.PeerID {
		delete(g.peers, oldAgent)
	}
	if existing, ok := g.peers[payload.PeerID]; ok && existing.agentID != agentID {
		log.Warn().Str("peer", payload.PeerID).Str("agent", agentID).Msg("P2P register: peer_id already taken")
		return
	}

	g.peers[payload.PeerID] = &registeredPeer{
		peerID:       payload.PeerID,
		session:      session,
		agentID:      agentID,
		registeredAt: time.Now().UTC(),
	}
	g.peerByAgent[agentID] = payload.PeerID
	log.Info().Str("peer", payload.PeerID).Str("agent", agentID).Msg("P2P peer registered")
}

func (g *tunnelGateway) handleP2PSignal(agentID string, session tunnel.Session, msg *tunnel.Message) {
	var sig signaling.Message
	if err := json.Unmarshal(msg.Payload, &sig); err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("P2P signal: invalid payload")
		return
	}
	if sig.To == "" {
		log.Warn().Str("agent", agentID).Msg("P2P signal: missing recipient")
		return
	}

	g.mu.RLock()
	senderPeer, senderOK := g.peerByAgent[agentID]
	if !senderOK {
		g.mu.RUnlock()
		log.Warn().Str("agent", agentID).Msg("P2P signal: sender not registered")
		return
	}
	if sig.From != "" && sig.From != senderPeer {
		g.mu.RUnlock()
		log.Warn().Str("agent", agentID).Str("from", sig.From).Str("registered", senderPeer).Msg("P2P signal: from mismatch")
		return
	}
	sig.From = senderPeer
	if !g.p2pSignalLimiter.Allow(senderPeer) {
		log.Warn().Str("from", senderPeer).Msg("P2P signal: rate limit exceeded")
		return
	}
	target, ok := g.peers[sig.To]
	g.mu.RUnlock()

	if !ok {
		log.Warn().Str("to", sig.To).Msg("P2P signal: peer not found")
		return
	}
	if g.authProvider != nil && sig.AuthToken != "" {
		if _, err := g.authProvider.ValidateToken(sig.AuthToken); err != nil {
			log.Warn().Err(err).Str("from", sig.From).Msg("P2P signal: auth failed")
			return
		}
	}

	payload, err := json.Marshal(sig)
	if err != nil {
		log.Warn().Err(err).Msg("P2P signal: marshal failed")
		return
	}
	g.deliverP2PSignal(target, payload)
}

func (g *tunnelGateway) deliverP2PSignal(target *registeredPeer, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := target.session.Open(ctx)
	if err != nil {
		log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: open stream to peer failed")
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: close stream failed")
		}
	}()

	if err := tunnel.WriteMessage(stream, &tunnel.Message{
		Type:    tunnel.MessageTypeP2PSignal,
		Payload: payload,
	}); err != nil {
		log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: write failed")
	}
}

func (g *tunnelGateway) removePeerLocked(agentID string) {
	peerID, ok := g.peerByAgent[agentID]
	if !ok {
		return
	}
	delete(g.peerByAgent, agentID)
	delete(g.peers, peerID)
	log.Info().Str("peer", peerID).Str("agent", agentID).Msg("P2P peer removed (agent disconnected)")
}

func (g *tunnelGateway) healthCheckLoop() {
	defer g.wg.Done()

	interval := time.Duration(g.cfg.HealthCheckInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopCh:
			return
		case <-ticker.C:
			g.checkServices()
		}
	}
}

func (g *tunnelGateway) checkServices() {
	// 心跳超时阈值
	heartbeatTimeout := time.Duration(g.cfg.HeartbeatTimeout) * time.Second
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 90 * time.Second
	}
	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	for name, svc := range g.services {
		if svc.session == nil || svc.session.IsClosed() {
			delete(g.services, name)
			log.Info().Str("service", name).Msg("Service removed (session closed)")
			continue
		}

		// 心跳超时检测：超过阈值未收到心跳则移除服务
		if !svc.info.LastHeartbeat.IsZero() && now.Sub(svc.info.LastHeartbeat) > heartbeatTimeout {
			svc.info.Status = tunnel.ServiceStatusOffline
			delete(g.services, name)
			log.Warn().
				Str("service", name).
				Dur("since_last_heartbeat", now.Sub(svc.info.LastHeartbeat)).
				Msg("Service removed (heartbeat timeout)")
			continue
		}

		// 检查服务健康状态
		status := g.checkServiceHealth(svc)
		if status != tunnel.ServiceStatusOnline {
			log.Warn().Str("service", name).Str("status", string(status)).Msg("Service health check failed")
		}
	}
}

func (g *tunnelGateway) checkServiceHealth(svc *registeredService) tunnel.ServiceStatus {
	// 检查会话状态
	if svc.session == nil || svc.session.IsClosed() {
		return tunnel.ServiceStatusOffline
	}

	// 检查每个端点的健康状态
	allHealthy := true
	for i, endpoint := range svc.info.Endpoints {
		// 确保 Metadata 已初始化，避免对 nil map 写入导致 panic
		if svc.info.Endpoints[i].Metadata == nil {
			svc.info.Endpoints[i].Metadata = make(map[string]string)
		}
		if !g.checkEndpointHealth(svc, &endpoint) {
			allHealthy = false
			log.Warn().Str("service", svc.info.Name).Str("endpoint", string(endpoint.Type)).Str("address", endpoint.Address).Msg("Endpoint health check failed")
			// 更新端点健康状态
			svc.info.Endpoints[i].Metadata["health_status"] = "unhealthy"
		} else {
			svc.info.Endpoints[i].Metadata["health_status"] = "healthy"
		}
	}

	if allHealthy {
		svc.info.Status = tunnel.ServiceStatusOnline
		return tunnel.ServiceStatusOnline
	} else {
		svc.info.Status = tunnel.ServiceStatusUnhealthy
		return tunnel.ServiceStatusUnhealthy
	}
}

func (g *tunnelGateway) checkEndpointHealth(svc *registeredService, endpoint *tunnel.Endpoint) bool {
	// 创建健康检查上下文
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 根据端点类型进行不同的健康检查
	switch endpoint.Type {
	case tunnel.EndpointTypeHTTP:
		return g.checkHTTPEndpointHealth(ctx, svc, endpoint)
	case tunnel.EndpointTypeGRPC:
		return g.checkGRPCEndpointHealth(ctx, svc, endpoint)
	case tunnel.EndpointTypeDebug:
		return g.checkDebugEndpointHealth(ctx, svc, endpoint)
	default:
		// 其他类型端点默认认为健康
		return true
	}
}

func (g *tunnelGateway) checkHTTPEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	// 打开到 Agent 的 stream
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return false
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close health check stream")
		}
	}()

	// 构建健康检查请求
	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: tunnel.EndpointTypeHTTP,
		Path:         endpoint.Path + "/health", // 假设健康检查路径为 /health
		Method:       "GET",
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		log.Warn().Err(err).Str("service", svc.info.Name).Msg("Gateway: failed to marshal health check meta")
		return false
	}

	// 发送健康检查请求
	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeHTTPRequest,
		Payload: payload,
	}
	if err := g.sendMessage(stream, msg); err != nil {
		return false
	}

	// 发送 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.Address+"/health", nil)
	if err != nil {
		return false
	}
	if err := req.Write(stream); err != nil {
		return false
	}

	// 读取响应
	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close health check response body")
		}
	}()

	// 检查响应状态码
	return resp.StatusCode == http.StatusOK
}

func (g *tunnelGateway) checkGRPCEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	// gRPC 健康检查实现
	// 这里简化处理，实际应该实现 gRPC 健康检查协议
	return true
}

func (g *tunnelGateway) checkDebugEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	// Debug 端点健康检查
	// 简化处理，检查是否能打开流
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return false
	}
	if err := stream.Close(); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to close debug health check stream")
		return false
	}
	return true
}

// FiberHandler returns a Fiber handler for the gateway HTTP proxy
func (g *tunnelGateway) FiberHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		serviceName := c.Params("service")
		if serviceName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "service name required",
			})
		}

		// Extract sub path after service name
		fullPath := c.Path()
		subPath := ""
		if idx := strings.Index(fullPath, serviceName); idx >= 0 {
			subPath = fullPath[idx+len(serviceName):]
		}

		g.mu.RLock()
		svc, ok := g.services[serviceName]
		g.mu.RUnlock()

		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "service not found",
			})
		}

		if svc.session == nil || svc.session.IsClosed() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "service unavailable",
			})
		}

		// Open a stream to the agent
		ctx := c.Context()
		stream, err := svc.session.Open(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to open stream: %v", err),
			})
		}
		defer func() {
			if err := stream.Close(); err != nil {
				log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
			}
		}()

		// Build request meta
		meta := tunnel.RequestMeta{
			ServiceID:    svc.info.ID,
			EndpointType: tunnel.EndpointTypeHTTP,
			Path:         subPath,
			Method:       c.Method(),
		}
		payload, err := json.Marshal(meta)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to build request meta: %v", err),
			})
		}

		// Send HTTP request message
		msg := &tunnel.Message{
			Type:    tunnel.MessageTypeHTTPRequest,
			Payload: payload,
		}

		if err := g.sendMessage(stream, msg); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to send message: %v", err),
			})
		}

		// Write the original request
		req := c.Request()
		if _, err := stream.Write(req.Header.Header()); err != nil {
			return err
		}
		if _, err := stream.Write(req.Body()); err != nil {
			return err
		}

		// Read response
		buf := make([]byte, 32*1024)
		for {
			n, err := stream.Read(buf)
			if n > 0 {
				if _, writeErr := c.Response().BodyWriter().Write(buf[:n]); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}

		return nil
	}
}

// ServeHTTP implements http.Handler for the gateway status
func (g *tunnelGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	services := g.Services()

	status := map[string]any{
		"status":       g.status.String(),
		"listen_addr":  g.cfg.ListenAddr,
		"transport":    g.cfg.Transport,
		"num_services": len(services),
		"services":     services,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to encode status")
	}
}
