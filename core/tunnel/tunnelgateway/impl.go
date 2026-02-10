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

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
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
	return &tunnelGateway{
		cfg:         cfg,
		services:    make(map[string]*registeredService),
		status:      tunnel.GatewayStatusStopped,
		rateLimiter: NewRateLimiter(100), // 默认每秒100个请求
	}
}

type registeredService struct {
	info    *tunnel.ServiceInfo
	session tunnel.Session
	agent   string // agent identifier
}

// RateLimiter 速率限制器
type RateLimiter struct {
	limits       map[string]int          // 服务名 -> 每秒最大请求数
	buckets      map[string]*TokenBucket // 服务名 -> 令牌桶
	mu           sync.RWMutex
	defaultLimit int // 默认速率限制
}

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity   int       // 令牌桶容量
	rate       int       // 每秒生成令牌数
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次填充时间
	mu         sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(defaultLimit int) *RateLimiter {
	return &RateLimiter{
		limits:       make(map[string]int),
		buckets:      make(map[string]*TokenBucket),
		defaultLimit: defaultLimit,
	}
}

// SetLimit 设置服务的速率限制
func (rl *RateLimiter) SetLimit(serviceName string, limit int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limits[serviceName] = limit
	// 重新创建令牌桶
	rl.buckets[serviceName] = NewTokenBucket(limit, limit)
}

// GetLimit 获取服务的速率限制
func (rl *RateLimiter) GetLimit(serviceName string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if limit, ok := rl.limits[serviceName]; ok {
		return limit
	}
	return rl.defaultLimit
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(serviceName string) bool {
	rl.mu.RLock()
	bucket, ok := rl.buckets[serviceName]
	if !ok {
		limit := rl.defaultLimit
		if l, ok := rl.limits[serviceName]; ok {
			limit = l
		}
		bucket = NewTokenBucket(limit, limit)
		rl.buckets[serviceName] = bucket
	}
	rl.mu.RUnlock()
	return bucket.Allow()
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算从上次填充到现在应该生成的令牌数
	now := time.Now()
	timeElapsed := now.Sub(tb.lastRefill).Seconds()
	tokensToAdd := timeElapsed * float64(tb.rate)

	// 填充令牌
	tb.tokens += tokensToAdd
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.lastRefill = now

	// 检查是否有足够的令牌
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

type tunnelGateway struct {
	cfg          *tunnel.GatewayConfig
	transport    tunnel.Transport
	listener     tunnel.Listener
	services     map[string]*registeredService
	status       tunnel.GatewayStatus
	authProvider tunnel.AuthProvider
	rateLimiter  *RateLimiter

	// 对外代理服务器
	httpServer  *http.Server
	debugServer *http.Server

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

		parts := strings.SplitN(path[1:], "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			// 没有指定服务名，返回服务列表
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
	json.NewEncoder(w).Encode(map[string]any{
		"services": services,
		"count":    len(services),
	})
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
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
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
	payload, _ := json.Marshal(meta)

	// 发送请求消息给 Agent
	msg := &tunnel.Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Payload: payload,
	}

	if err := g.sendMessage(stream, msg); err != nil {
		stream.Close()
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
	defer stream.Close()

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
			resp.Body.Close()
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
		stream.Close()
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		stream.Close()
		log.Warn().Err(err).Str("service", serviceName).Msg("Failed to hijack connection")
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}

	// 将原始 HTTP 请求写入 stream，让 Agent 处理 WebSocket 升级
	if err := r.Write(stream); err != nil {
		clientConn.Close()
		stream.Close()
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
		io.Copy(stream, clientConn)
		stream.Close()
	}()

	// Agent -> Client
	go func() {
		defer wg.Done()
		io.Copy(clientConn, stream)
		clientConn.Close()
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
		g.httpServer.Shutdown(shutdownCtx)
	}
	if g.debugServer != nil {
		g.debugServer.Shutdown(shutdownCtx)
	}

	// Close listener
	if g.listener != nil {
		g.listener.Close()
	}

	// Close all sessions
	g.mu.Lock()
	for _, svc := range g.services {
		if svc.session != nil {
			svc.session.Close()
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
	defer stream.Close()

	// Build request meta
	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: endpointType,
	}
	payload, _ := json.Marshal(meta)

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
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write length prefix (4 bytes) + data
	length := uint32(len(data))
	header := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	if _, err := stream.Write(header); err != nil {
		return err
	}
	if _, err := stream.Write(data); err != nil {
		return err
	}
	return nil
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
	defer stream.Close()

	// Read message header
	header := make([]byte, 4)
	if _, err := io.ReadFull(stream, header); err != nil {
		log.Warn().Err(err).Msg("Failed to read message header")
		return
	}

	length := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Warn().Err(err).Msg("Failed to read message data")
		return
	}

	var msg tunnel.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal message")
		return
	}

	switch msg.Type {
	case tunnel.MessageTypeRegister:
		g.handleRegister(agentID, session, &msg)
	case tunnel.MessageTypeDeregister:
		g.handleDeregister(&msg)
	case tunnel.MessageTypeHeartbeat:
		g.handleHeartbeat(agentID)
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

	g.mu.Lock()
	g.services[service.Name] = &registeredService{
		info:    &service,
		session: session,
		agent:   agentID,
	}
	g.mu.Unlock()

	log.Info().Str("service", service.Name).Str("agent", agentID).Msg("Service registered")
}

func (g *tunnelGateway) handleDeregister(msg *tunnel.Message) {
	if len(msg.Payload) == 0 {
		return
	}

	serviceName := string(msg.Payload)

	g.mu.Lock()
	delete(g.services, serviceName)
	g.mu.Unlock()

	log.Info().Str("service", serviceName).Msg("Service deregistered")
}

func (g *tunnelGateway) handleHeartbeat(agentID string) {
	// Update last heartbeat time (could be used for health checking)
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
	g.mu.Lock()
	defer g.mu.Unlock()

	for name, svc := range g.services {
		if svc.session == nil || svc.session.IsClosed() {
			delete(g.services, name)
			log.Info().Str("service", name).Msg("Service removed (session closed)")
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
	defer stream.Close()

	// 构建健康检查请求
	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: tunnel.EndpointTypeHTTP,
		Path:         endpoint.Path + "/health", // 假设健康检查路径为 /health
		Method:       "GET",
	}
	payload, _ := json.Marshal(meta)

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
	defer resp.Body.Close()

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
	stream.Close()
	return true
}

// FiberHandler returns a Fiber handler for the gateway HTTP proxy
func (g *tunnelGateway) FiberHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
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
		ctx := c.UserContext()
		stream, err := svc.session.Open(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to open stream: %v", err),
			})
		}
		defer stream.Close()

		// Build request meta
		meta := tunnel.RequestMeta{
			ServiceID:    svc.info.ID,
			EndpointType: tunnel.EndpointTypeHTTP,
			Path:         subPath,
			Method:       c.Method(),
		}
		payload, _ := json.Marshal(meta)

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
				c.Response().BodyWriter().Write(buf[:n])
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
	json.NewEncoder(w).Encode(status)
}
