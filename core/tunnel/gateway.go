package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/log"
)

var _ Gateway = (*tunnelGateway)(nil)

// NewGateway creates a new tunnel gateway
func NewGateway(cfg *GatewayConfig) Gateway {
	return &tunnelGateway{
		cfg:      cfg,
		services: make(map[string]*registeredService),
		status:   GatewayStatusStopped,
	}
}

type registeredService struct {
	info    *ServiceInfo
	session Session
	agent   string // agent identifier
}

type tunnelGateway struct {
	cfg       *GatewayConfig
	transport Transport
	listener  Listener
	services  map[string]*registeredService
	status    GatewayStatus

	// 对外代理服务器
	httpServer  *http.Server
	debugServer *http.Server

	mu       sync.RWMutex
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	running  atomic.Bool
}

func (g *tunnelGateway) Start(ctx context.Context) error {
	if g.running.Load() {
		return ErrGatewayAlreadyRunning
	}

	// Create transport
	transport, err := NewTransport(g.cfg.Transport, g.cfg.TransportOptions)
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
	g.status = GatewayStatusRunning

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
		Handler: g.createProxyHandler(EndpointTypeHTTP),
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
		Handler: g.createProxyHandler(EndpointTypeDebug),
	}

	log.Info().Str("addr", addr).Msg("Debug proxy server started")

	if err := g.debugServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Debug proxy server error")
	}
}

// createProxyHandler 创建 HTTP 代理处理器
// URL 格式: /{service_name}/path... -> 转发到对应 Agent 的本地服务
func (g *tunnelGateway) createProxyHandler(endpointType EndpointType) http.Handler {
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
	services := make([]*ServiceInfo, 0, len(g.services))
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
func (g *tunnelGateway) proxyToAgent(w http.ResponseWriter, r *http.Request, svc *registeredService, endpointType EndpointType, subPath string) {
	ctx := r.Context()

	// 打开到 Agent 的 stream
	stream, err := svc.session.Open(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open stream: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// 发送请求消息给 Agent
	msg := &Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Service: svc.info,
	}

	if err := g.sendMessage(stream, msg); err != nil {
		http.Error(w, fmt.Sprintf("failed to send message: %v", err), http.StatusInternalServerError)
		return
	}

	// 修改请求路径为子路径
	r.URL.Path = subPath
	r.RequestURI = subPath
	if r.URL.RawQuery != "" {
		r.RequestURI = subPath + "?" + r.URL.RawQuery
	}

	// 使用 httputil 进行双向代理
	// 创建一个虚拟的后端连接
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// 保持原始请求
		},
		Transport: &streamRoundTripper{stream: stream, request: r},
	}

	proxy.ServeHTTP(w, r)
}

// streamRoundTripper 实现 http.RoundTripper，通过 stream 转发请求
type streamRoundTripper struct {
	stream  Stream
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
	g.status = GatewayStatusStopped
	return nil
}

func (g *tunnelGateway) Services() []*ServiceInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(g.services))
	for _, svc := range g.services {
		services = append(services, svc.info)
	}
	return services
}

func (g *tunnelGateway) GetService(name string) (*ServiceInfo, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	svc, ok := g.services[name]
	if !ok {
		return nil, ErrServiceNotFound
	}
	return svc.info, nil
}

func (g *tunnelGateway) Status() GatewayStatus {
	return g.status
}

func (g *tunnelGateway) Forward(ctx context.Context, serviceName string, endpointType EndpointType, conn net.Conn) error {
	g.mu.RLock()
	svc, ok := g.services[serviceName]
	g.mu.RUnlock()

	if !ok {
		return ErrServiceNotFound
	}

	if svc.session == nil || svc.session.IsClosed() {
		return ErrSessionClosed
	}

	// Open a stream to the agent
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	// Send forward request
	msg := &Message{
		Type:    g.endpointTypeToMessageType(endpointType),
		Service: svc.info,
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

func (g *tunnelGateway) endpointTypeToMessageType(et EndpointType) MessageType {
	switch et {
	case EndpointTypeHTTP:
		return MessageTypeHTTPRequest
	case EndpointTypeGRPC:
		return MessageTypeGRPCRequest
	case EndpointTypeDebug:
		return MessageTypeDebugRequest
	default:
		return MessageTypeHTTPRequest
	}
}

func (g *tunnelGateway) sendMessage(stream Stream, msg *Message) error {
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

func (g *tunnelGateway) handleSession(session Session) {
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

func (g *tunnelGateway) handleStream(agentID string, session Session, stream Stream) {
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

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal message")
		return
	}

	switch msg.Type {
	case MessageTypeRegister:
		g.handleRegister(agentID, session, &msg)
	case MessageTypeDeregister:
		g.handleDeregister(&msg)
	case MessageTypeHeartbeat:
		g.handleHeartbeat(agentID)
	}
}

func (g *tunnelGateway) handleRegister(agentID string, session Session, msg *Message) {
	if msg.Service == nil {
		return
	}

	g.mu.Lock()
	g.services[msg.Service.Name] = &registeredService{
		info:    msg.Service,
		session: session,
		agent:   agentID,
	}
	g.mu.Unlock()

	log.Info().Str("service", msg.Service.Name).Str("agent", agentID).Msg("Service registered")
}

func (g *tunnelGateway) handleDeregister(msg *Message) {
	if msg.Service == nil {
		return
	}

	g.mu.Lock()
	delete(g.services, msg.Service.Name)
	g.mu.Unlock()

	log.Info().Str("service", msg.Service.Name).Msg("Service deregistered")
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
		}
	}
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

		// Send HTTP request message
		msg := &Message{
			Type:    MessageTypeHTTPRequest,
			Service: svc.info,
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
