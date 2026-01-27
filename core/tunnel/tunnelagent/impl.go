package tunnelagent

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pubgo/funk/v2/log"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/core/tunnel"
)

var _ tunnel.Agent = (*tunnelAgent)(nil)

// NewAgent creates a new tunnel agent
func NewAgent(cfg *tunnel.AgentConfig) tunnel.Agent {
	a := &tunnelAgent{
		cfg:      cfg,
		services: make(map[string]*tunnel.ServiceInfo),
		status:   tunnel.StatusDisconnected,
	}

	// 从配置中构建初始服务信息
	if cfg.ServiceName != "" {
		svc := &tunnel.ServiceInfo{
			ID:       cfg.ServiceID,
			Name:     cfg.ServiceName,
			Version:  cfg.ServiceVersion,
			Metadata: cfg.Metadata,
		}

		// 转换 Endpoints 配置到 ServiceInfo.Endpoints
		for _, ep := range cfg.Endpoints {
			svc.Endpoints = append(svc.Endpoints, tunnel.Endpoint{
				Type:     tunnel.EndpointType(ep.Type),
				Address:  ep.LocalAddr,
				Path:     ep.Path,
				Metadata: ep.Metadata,
			})
		}

		a.services[cfg.ServiceName] = svc
	}

	return a
}

type tunnelAgent struct {
	cfg       *tunnel.AgentConfig
	transport tunnel.Transport
	session   tunnel.Session
	services  map[string]*tunnel.ServiceInfo
	status    tunnel.AgentStatus

	mu       sync.RWMutex
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	running  atomic.Bool

	// Reverse proxies for forwarding requests to local services
	httpProxy *httputil.ReverseProxy
	grpcConn  *grpc.ClientConn
}

func (a *tunnelAgent) Start(ctx context.Context) error {
	if a.running.Load() {
		return tunnel.ErrAgentAlreadyRunning
	}

	// Create transport
	transport, err := tunnel.NewTransport(a.cfg.Transport, a.cfg.TransportOptions)
	if err != nil {
		return err
	}
	a.transport = transport

	// Connect to gateway
	if err := a.connect(ctx); err != nil {
		return err
	}

	a.stopCh = make(chan struct{})
	a.running.Store(true)
	a.status = tunnel.StatusConnected

	// Start heartbeat
	a.wg.Add(1)
	go a.heartbeatLoop()

	// Start accepting streams from gateway
	a.wg.Add(1)
	go a.acceptLoop()

	// Start reconnection monitor
	a.wg.Add(1)
	go a.reconnectLoop(ctx)

	return nil
}

func (a *tunnelAgent) Stop(ctx context.Context) error {
	if !a.running.Load() {
		return nil
	}

	a.stopOnce.Do(func() {
		close(a.stopCh)
	})

	// 先关闭 session，让所有等待的 goroutine 退出
	if a.session != nil {
		a.session.Close()
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		// 超时但继续
		log.Warn().Msg("Agent stop timeout, force closing")
	}

	a.running.Store(false)
	a.status = tunnel.StatusDisconnected

	return nil
}

func (a *tunnelAgent) Register(ctx context.Context, service *tunnel.ServiceInfo) error {
	a.mu.Lock()
	a.services[service.Name] = service
	a.mu.Unlock()

	if a.session == nil || a.session.IsClosed() {
		return nil // Will register when connected
	}

	return a.sendRegister(ctx, service)
}

func (a *tunnelAgent) Deregister(ctx context.Context, serviceName string) error {
	a.mu.Lock()
	delete(a.services, serviceName)
	a.mu.Unlock()

	if a.session == nil || a.session.IsClosed() {
		return nil
	}

	return a.sendDeregister(ctx, serviceName)
}

func (a *tunnelAgent) Status() tunnel.AgentStatus {
	return a.status
}

func (a *tunnelAgent) Info() *tunnel.AgentInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	info := &tunnel.AgentInfo{
		GatewayAddr:    a.cfg.GatewayAddr,
		ServiceName:    a.cfg.ServiceName,
		ServiceVersion: a.cfg.ServiceVersion,
		Status:         a.status.String(),
	}

	// 收集端点信息
	for _, svc := range a.services {
		info.Endpoints = append(info.Endpoints, svc.Endpoints...)
	}

	return info
}

func (a *tunnelAgent) connect(ctx context.Context) error {
	session, err := a.transport.Dial(ctx, a.cfg.GatewayAddr)
	if err != nil {
		a.status = tunnel.StatusDisconnected
		return err
	}
	a.session = session
	a.status = tunnel.StatusConnected

	// Register all services
	a.mu.RLock()
	services := make([]*tunnel.ServiceInfo, 0, len(a.services))
	for _, svc := range a.services {
		services = append(services, svc)
	}
	a.mu.RUnlock()

	for _, svc := range services {
		if err := a.sendRegister(ctx, svc); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("Failed to register service")
		}
	}

	return nil
}

func (a *tunnelAgent) sendRegister(ctx context.Context, service *tunnel.ServiceInfo) error {
	stream, err := a.session.Open(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	payload, err := json.Marshal(service)
	if err != nil {
		return err
	}

	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeRegister,
		Payload: payload,
	}
	return a.sendMessage(stream, msg)
}

func (a *tunnelAgent) sendDeregister(ctx context.Context, serviceName string) error {
	stream, err := a.session.Open(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeDeregister,
		Payload: []byte(serviceName),
	}
	return a.sendMessage(stream, msg)
}

func (a *tunnelAgent) sendMessage(stream tunnel.Stream, msg *tunnel.Message) error {
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

func (a *tunnelAgent) heartbeatLoop() {
	defer a.wg.Done()

	interval := time.Duration(a.cfg.HeartbeatInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Warn().Err(err).Msg("Failed to send heartbeat")
			}
		}
	}
}

func (a *tunnelAgent) sendHeartbeat() error {
	if a.session == nil || a.session.IsClosed() {
		return tunnel.ErrSessionClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := a.session.Open(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	msg := &tunnel.Message{Type: tunnel.MessageTypeHeartbeat}
	return a.sendMessage(stream, msg)
}

func (a *tunnelAgent) acceptLoop() {
	defer a.wg.Done()

	log.Debug().Msg("Agent: acceptLoop started, waiting for streams from gateway")

	for {
		select {
		case <-a.stopCh:
			log.Debug().Msg("Agent: acceptLoop stopped by stopCh")
			return
		default:
		}

		if a.session == nil || a.session.IsClosed() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		stream, err := a.session.Accept()
		if err != nil {
			if a.session.IsClosed() {
				log.Debug().Msg("Agent: acceptLoop stopped, session closed")
				return
			}
			log.Warn().Err(err).Msg("Agent: Failed to accept stream from gateway")
			continue
		}

		log.Debug().Msg("Agent: Accepted stream from gateway, handling...")
		go a.handleStream(stream)
	}
}

func (a *tunnelAgent) handleStream(stream tunnel.Stream) {
	log.Debug().Msg("Agent: Accepted new stream from gateway")

	// Read message header
	header := make([]byte, 4)
	if _, err := io.ReadFull(stream, header); err != nil {
		log.Warn().Err(err).Msg("Agent: Failed to read message header")
		stream.Close()
		return
	}

	length := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
	log.Debug().Uint32("length", length).Msg("Agent: Read message header")

	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Warn().Err(err).Msg("Agent: Failed to read message data")
		stream.Close()
		return
	}

	var msg tunnel.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Err(err).Str("data", string(data)).Msg("Agent: Failed to unmarshal message")
		stream.Close()
		return
	}

	log.Debug().Str("type", string(msg.Type)).Msg("Received message from gateway")

	switch msg.Type {
	case tunnel.MessageTypeHTTPRequest:
		a.handleHTTPRequest(stream, &msg)
	case tunnel.MessageTypeGRPCRequest:
		a.handleGRPCRequest(stream, &msg)
	case tunnel.MessageTypeDebugRequest:
		a.handleDebugRequest(stream, &msg)
	default:
		stream.Close()
	}
}

func (a *tunnelAgent) handleHTTPRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer stream.Close()

	// Parse request meta from payload
	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("HTTP request: failed to parse request meta")
			return
		}
	}

	// Find the HTTP endpoint from local services
	var httpEndpoint *tunnel.Endpoint
	a.mu.RLock()
	for _, svc := range a.services {
		for i := range svc.Endpoints {
			if svc.Endpoints[i].Type == tunnel.EndpointTypeHTTP {
				httpEndpoint = &svc.Endpoints[i]
				break
			}
		}
		if httpEndpoint != nil {
			break
		}
	}
	a.mu.RUnlock()

	if httpEndpoint == nil {
		log.Warn().Msg("HTTP request: no HTTP endpoint found")
		return
	}

	// Normalize address: add localhost if address starts with ':'
	address := httpEndpoint.Address
	if len(address) > 0 && address[0] == ':' {
		address = "127.0.0.1" + address
	}

	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying HTTP request to local service")

	// Forward request to local HTTP service
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Failed to connect to local HTTP service")
		return
	}
	defer conn.Close()

	// Bidirectional copy - wait for both directions to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// stream -> conn (request from gateway to local service)
	go func() {
		defer wg.Done()
		io.Copy(conn, stream)
		// Close write side to signal end of request
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// conn -> stream (response from local service to gateway)
	go func() {
		defer wg.Done()
		io.Copy(stream, conn)
	}()

	wg.Wait()
	log.Debug().Str("address", address).Msg("HTTP request completed")
}

func (a *tunnelAgent) handleGRPCRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer stream.Close()

	// Parse request meta from payload
	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("gRPC request: failed to parse request meta")
			return
		}
	}

	// Find the gRPC endpoint from local services
	var grpcEndpoint *tunnel.Endpoint
	a.mu.RLock()
	for _, svc := range a.services {
		for i := range svc.Endpoints {
			if svc.Endpoints[i].Type == tunnel.EndpointTypeGRPC {
				grpcEndpoint = &svc.Endpoints[i]
				break
			}
		}
		if grpcEndpoint != nil {
			break
		}
	}
	a.mu.RUnlock()

	if grpcEndpoint == nil {
		log.Warn().Msg("gRPC request: no gRPC endpoint found")
		return
	}

	// Normalize address: add localhost if address starts with ':'
	address := grpcEndpoint.Address
	if len(address) > 0 && address[0] == ':' {
		address = "127.0.0.1" + address
	}

	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying gRPC request to local service")

	// Forward request to local gRPC service
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Failed to connect to local gRPC service")
		return
	}
	defer conn.Close()

	// Bidirectional copy - wait for both directions to complete
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(conn, stream)
	}()

	go func() {
		defer wg.Done()
		io.Copy(stream, conn)
	}()

	wg.Wait()
	log.Debug().Str("address", address).Msg("gRPC request completed")
}

func (a *tunnelAgent) handleDebugRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer stream.Close()

	// Parse request meta from payload
	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("Debug request: failed to parse request meta")
			return
		}
	}

	// Find the debug endpoint from local services
	var debugEndpoint *tunnel.Endpoint
	a.mu.RLock()
	for _, svc := range a.services {
		for i := range svc.Endpoints {
			if svc.Endpoints[i].Type == tunnel.EndpointTypeDebug {
				debugEndpoint = &svc.Endpoints[i]
				break
			}
		}
		if debugEndpoint != nil {
			break
		}
	}
	a.mu.RUnlock()

	if debugEndpoint == nil {
		log.Warn().Msg("Debug request: no debug endpoint found")
		return
	}

	// Normalize address: add localhost if address starts with ':'
	address := debugEndpoint.Address
	if len(address) > 0 && address[0] == ':' {
		address = "127.0.0.1" + address
	}

	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying debug request to local service")

	// Forward request to local debug service
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Failed to connect to local debug service")
		return
	}
	defer conn.Close()

	// Bidirectional copy - wait for both directions to complete
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(conn, stream)
	}()

	go func() {
		defer wg.Done()
		io.Copy(stream, conn)
	}()

	wg.Wait()
	log.Debug().Str("address", address).Msg("Debug request completed")
}

func (a *tunnelAgent) reconnectLoop(ctx context.Context) {
	defer a.wg.Done()

	interval := time.Duration(a.cfg.ReconnectInterval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	for {
		select {
		case <-a.stopCh:
			return
		case <-time.After(interval):
			if a.session == nil || a.session.IsClosed() {
				a.status = tunnel.StatusReconnecting
				if err := a.connect(ctx); err != nil {
					log.Warn().Err(err).Msg("Failed to reconnect to gateway")
				}
			}
		}
	}
}

// ServeHTTP implements http.Handler for the agent status
func (a *tunnelAgent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	services := make([]*tunnel.ServiceInfo, 0, len(a.services))
	for _, svc := range a.services {
		services = append(services, svc)
	}
	a.mu.RUnlock()

	status := map[string]any{
		"status":      a.status.String(),
		"gateway":     a.cfg.GatewayAddr,
		"transport":   a.cfg.Transport,
		"services":    services,
		"num_streams": 0,
	}

	if a.session != nil && !a.session.IsClosed() {
		status["num_streams"] = a.session.NumStreams()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
