package tunnelagent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	if cfg != nil {
		cfg.Normalize()
	}
	a := &tunnelAgent{
		cfg:      cfg,
		services: make(map[string]*tunnel.ServiceInfo),
		status:   tunnel.StatusDisconnected,
		stats: &Stats{
			ServiceStats: make(map[string]*ServiceStats),
			LastActivity: time.Now(),
		},
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
		a.stats.ServiceStats[cfg.ServiceName] = &ServiceStats{}
	}

	return a
}

// Stats 统计信息
type Stats struct {
	Connections   int64                    `json:"connections"`
	Streams       int64                    `json:"streams"`
	BytesIn       int64                    `json:"bytes_in"`
	BytesOut      int64                    `json:"bytes_out"`
	Requests      int64                    `json:"requests"`
	Errors        int64                    `json:"errors"`
	ResponseTimes []time.Duration          `json:"response_times"`
	ServiceStats  map[string]*ServiceStats `json:"service_stats"`
	LastActivity  time.Time                `json:"last_activity"`
}

// ServiceStats 服务统计信息
type ServiceStats struct {
	Requests      int64           `json:"requests"`
	Errors        int64           `json:"errors"`
	BytesIn       int64           `json:"bytes_in"`
	BytesOut      int64           `json:"bytes_out"`
	ResponseTimes []time.Duration `json:"response_times"`
	LastRequest   time.Time       `json:"last_request"`
}

type tunnelAgent struct {
	cfg       *tunnel.AgentConfig
	transport tunnel.Transport
	session   tunnel.Session
	services  map[string]*tunnel.ServiceInfo
	status    tunnel.AgentStatus
	stats     *Stats
	statsMu   sync.Mutex

	mu           sync.RWMutex
	connectMu    sync.Mutex  // 串行化重连，避免并发 connect 产生多个 session
	reconnecting atomic.Bool // 防止立即重连 goroutine 堆积
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
	running      atomic.Bool

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
		if err := a.session.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close session")
		}
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

// Session 返回当前与 gateway 的隧道连接（已连接时非 nil）。
func (a *tunnelAgent) Session() tunnel.Session {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.session
}

func (a *tunnelAgent) connect(ctx context.Context) error {
	session, err := a.transport.Dial(ctx, a.cfg.GatewayAddr)
	if err != nil {
		a.status = tunnel.StatusDisconnected
		atomic.AddInt64(&a.stats.Errors, 1)
		a.updateLastActivity()
		return err
	}
	a.mu.Lock()
	a.session = session
	a.mu.Unlock()
	a.status = tunnel.StatusConnected

	// 更新统计信息
	atomic.AddInt64(&a.stats.Connections, 1)
	a.updateLastActivity()

	// Register all services
	a.mu.RLock()
	services := make([]*tunnel.ServiceInfo, 0, len(a.services))
	for _, svc := range a.services {
		services = append(services, svc)
		// 确保服务统计信息存在
		a.getOrCreateServiceStats(svc.Name)
	}
	a.mu.RUnlock()

	for _, svc := range services {
		if err := a.sendRegister(ctx, svc); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("Failed to register service")
			stats := a.getOrCreateServiceStats(svc.Name)
			atomic.AddInt64(&stats.Errors, 1)
		}
	}

	return nil
}

func (a *tunnelAgent) sendRegister(ctx context.Context, service *tunnel.ServiceInfo) error {
	// 服务注册使用高优先级
	stream, err := a.session.OpenWithPriority(ctx, 2)
	if err != nil {
		// 降级到普通优先级
		stream, err = a.session.Open(ctx)
		if err != nil {
			a.handleError(err)
			return err
		}
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close register stream")
		}
	}()

	payload, err := json.Marshal(service)
	if err != nil {
		return err
	}

	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeRegister,
		Payload: payload,
	}
	if err := a.sendMessage(stream, msg); err != nil {
		a.handleError(err)
		return err
	}
	return nil
}

func (a *tunnelAgent) sendDeregister(ctx context.Context, serviceName string) error {
	// 服务注销使用高优先级
	stream, err := a.session.OpenWithPriority(ctx, 2)
	if err != nil {
		// 降级到普通优先级
		stream, err = a.session.Open(ctx)
		if err != nil {
			a.handleError(err)
			return err
		}
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close deregister stream")
		}
	}()

	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeDeregister,
		Payload: []byte(serviceName),
	}
	if err := a.sendMessage(stream, msg); err != nil {
		a.handleError(err)
		return err
	}
	return nil
}

func (a *tunnelAgent) sendMessage(stream tunnel.Stream, msg *tunnel.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		atomic.AddInt64(&a.stats.Errors, 1)
		a.updateLastActivity()
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
		atomic.AddInt64(&a.stats.Errors, 1)
		a.updateLastActivity()
		return err
	}
	if _, err := stream.Write(data); err != nil {
		atomic.AddInt64(&a.stats.Errors, 1)
		a.updateLastActivity()
		return err
	}

	// 更新统计信息
	atomic.AddInt64(&a.stats.BytesOut, int64(len(header)+len(data)))
	a.updateLastActivity()
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

func (a *tunnelAgent) handleError(err error) {
	switch {
	case errors.Is(err, tunnel.ErrSessionClosed):
		// 会话关闭，需要重连
		log.Warn().Err(err).Msg("Session closed, will reconnect")
		go a.reconnectImmediately()
	case errors.Is(err, tunnel.ErrConnectionFailed):
		// 连接失败，需要重连
		log.Warn().Err(err).Msg("Connection failed, will reconnect")
		go a.reconnectImmediately()
	case errors.Is(err, tunnel.ErrTimeout):
		// 超时错误，可能是网络问题，需要重连
		log.Warn().Err(err).Msg("Timeout error, will reconnect")
		go a.reconnectImmediately()
	default:
		// 其他错误，记录但不需要重连
		log.Warn().Err(err).Msg("Unexpected error")
	}
}

func (a *tunnelAgent) reconnectImmediately() {
	// 防止多个 goroutine 同时触发立即重连导致堆积
	if !a.reconnecting.CompareAndSwap(false, true) {
		return
	}
	defer a.reconnecting.Store(false)

	// 立即尝试重连，而不是等待重连计时器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.reconnect(ctx); err != nil {
		log.Warn().Err(err).Msg("Immediate reconnect failed")
	}
}

// reconnect 串行化的重连入口：若已有健康会话则直接返回，否则重新建立连接。
func (a *tunnelAgent) reconnect(ctx context.Context) error {
	a.connectMu.Lock()
	defer a.connectMu.Unlock()

	a.mu.RLock()
	sess := a.session
	a.mu.RUnlock()
	if sess != nil && !sess.IsClosed() {
		return nil
	}

	a.status = tunnel.StatusReconnecting
	return a.connect(ctx)
}

func (a *tunnelAgent) sendHeartbeat() error {
	if a.session == nil || a.session.IsClosed() {
		return tunnel.ErrSessionClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 心跳使用中优先级
	stream, err := a.session.OpenWithPriority(ctx, 5)
	if err != nil {
		// 降级到普通优先级
		stream, err = a.session.Open(ctx)
		if err != nil {
			a.handleError(err)
			return err
		}
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close heartbeat stream")
		}
	}()

	msg := &tunnel.Message{Type: tunnel.MessageTypeHeartbeat}
	if err := a.sendMessage(stream, msg); err != nil {
		a.handleError(err)
		return err
	}
	return nil
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
	startTime := time.Now()
	log.Debug().Msg("Agent: Accepted new stream from gateway")

	// 更新流统计信息
	atomic.AddInt64(&a.stats.Streams, 1)
	a.updateLastActivity()

	// Read message header
	header := make([]byte, 4)
	if _, err := io.ReadFull(stream, header); err != nil {
		log.Warn().Err(err).Msg("Agent: Failed to read message header")
		atomic.AddInt64(&a.stats.Errors, 1)
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close stream after header read error")
		}
		return
	}

	length := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])
	log.Debug().Uint32("length", length).Msg("Agent: Read message header")

	// 防止超大长度前缀导致 OOM
	if length > tunnel.MaxMessageSize {
		log.Warn().Uint32("length", length).Int("limit", tunnel.MaxMessageSize).Msg("Agent: message size exceeds limit")
		atomic.AddInt64(&a.stats.Errors, 1)
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close stream after oversize message")
		}
		return
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Warn().Err(err).Msg("Agent: Failed to read message data")
		atomic.AddInt64(&a.stats.Errors, 1)
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close stream after data read error")
		}
		return
	}

	// 更新入站流量统计
	atomic.AddInt64(&a.stats.BytesIn, int64(len(header)+len(data)))

	var msg tunnel.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Err(err).Str("data", string(data)).Msg("Agent: Failed to unmarshal message")
		atomic.AddInt64(&a.stats.Errors, 1)
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close stream after unmarshal error")
		}
		return
	}

	log.Debug().Str("type", string(msg.Type)).Msg("Received message from gateway")

	// 处理请求
	switch msg.Type {
	case tunnel.MessageTypeHTTPRequest:
		a.handleHTTPRequest(stream, &msg)
	case tunnel.MessageTypeGRPCRequest:
		a.handleGRPCRequest(stream, &msg)
	case tunnel.MessageTypeDebugRequest:
		a.handleDebugRequest(stream, &msg)
	case tunnel.MessageTypeP2PSignal:
		a.handleP2PSignal(&msg)
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close P2P signal stream")
		}
	default:
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close stream for unknown message")
		}
	}

	// 计算响应时间并更新统计信息
	responseTime := time.Since(startTime)
	atomic.AddInt64(&a.stats.Requests, 1)
	a.recordResponseTime(responseTime)
}

func (a *tunnelAgent) handleHTTPRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close HTTP stream")
		}
	}()

	// Parse request meta from payload
	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("HTTP request: failed to parse request meta")
			return
		}
	}

	// Find the HTTP endpoint from local services
	a.mu.RLock()
	httpEndpoint := findEndpoint(a.services, meta, tunnel.EndpointTypeHTTP)
	a.mu.RUnlock()

	if httpEndpoint == nil {
		log.Warn().Str("service_id", meta.ServiceID).Msg("HTTP request: no HTTP endpoint found")
		return
	}

	address := normalizeAddress(httpEndpoint.Address)

	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying HTTP request to local service")
	proxyStreamToTCP(stream, address)
}

func (a *tunnelAgent) handleGRPCRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close gRPC stream")
		}
	}()

	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("gRPC request: failed to parse request meta")
			return
		}
	}

	a.mu.RLock()
	grpcEndpoint := findEndpoint(a.services, meta, tunnel.EndpointTypeGRPC)
	a.mu.RUnlock()

	if grpcEndpoint == nil {
		log.Warn().Str("service_id", meta.ServiceID).Msg("gRPC request: no gRPC endpoint found")
		return
	}

	address := normalizeAddress(grpcEndpoint.Address)
	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying gRPC request to local service")
	proxyStreamToTCP(stream, address)
}

func (a *tunnelAgent) handleDebugRequest(stream tunnel.Stream, msg *tunnel.Message) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Agent: failed to close debug stream")
		}
	}()

	var meta tunnel.RequestMeta
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &meta); err != nil {
			log.Warn().Err(err).Msg("Debug request: failed to parse request meta")
			return
		}
	}

	a.mu.RLock()
	debugEndpoint := findEndpoint(a.services, meta, tunnel.EndpointTypeDebug)
	a.mu.RUnlock()

	if debugEndpoint == nil {
		log.Warn().Str("service_id", meta.ServiceID).Msg("Debug request: no debug endpoint found")
		return
	}

	address := normalizeAddress(debugEndpoint.Address)
	log.Debug().Str("address", address).Str("path", meta.Path).Msg("Proxying debug request to local service")
	proxyStreamToTCP(stream, address)
}

func (a *tunnelAgent) handleP2PSignal(msg *tunnel.Message) {
	if a.cfg.P2PSignalHandler == nil {
		log.Warn().Msg("Agent: P2P signal received but no handler configured")
		return
	}
	a.cfg.P2PSignalHandler(msg.Payload)
}

func (a *tunnelAgent) reconnectLoop(ctx context.Context) {
	defer a.wg.Done()

	baseInterval := time.Duration(a.cfg.ReconnectInterval) * time.Second
	if baseInterval <= 0 {
		baseInterval = 5 * time.Second
	}
	maxInterval := 60 * time.Second // 最大重连间隔
	currentInterval := baseInterval
	attempt := 0

	maxAttempts := a.cfg.MaxReconnectAttempts // 0 表示无限重试

	for {
		select {
		case <-a.stopCh:
			return
		case <-time.After(currentInterval):
			a.mu.RLock()
			sess := a.session
			a.mu.RUnlock()

			if sess != nil && !sess.IsClosed() {
				// 会话正常，重置退避计时器
				if attempt > 0 {
					currentInterval = baseInterval
					attempt = 0
				}
				continue
			}

			attempt++
			log.Info().Int("attempt", attempt).Dur("interval", currentInterval).Msg("Attempting to reconnect to gateway")
			if err := a.reconnect(ctx); err != nil {
				log.Warn().Err(err).Int("attempt", attempt).Dur("interval", currentInterval).Msg("Failed to reconnect to gateway")

				// 达到最大重连次数则放弃
				if maxAttempts > 0 && attempt >= maxAttempts {
					log.Error().Int("attempt", attempt).Int("max_attempts", maxAttempts).Msg("Max reconnect attempts reached, giving up")
					a.status = tunnel.StatusDisconnected
					return
				}

				// 指数退避：每次失败后间隔翻倍
				currentInterval *= 2
				if currentInterval > maxInterval {
					currentInterval = maxInterval
				}
			} else {
				// 重连成功，重置退避计时器和尝试次数
				log.Info().Int("attempt", attempt).Msg("Successfully reconnected to gateway")
				currentInterval = baseInterval
				attempt = 0
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
	statsSnapshot := a.snapshotStats()
	status["stats"] = statsSnapshot

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Warn().Err(err).Msg("Agent: failed to encode status")
	}
}

func (a *tunnelAgent) updateLastActivity() {
	a.statsMu.Lock()
	a.stats.LastActivity = time.Now()
	a.statsMu.Unlock()
}

func (a *tunnelAgent) recordResponseTime(d time.Duration) {
	a.statsMu.Lock()
	a.stats.ResponseTimes = append(a.stats.ResponseTimes, d)
	if len(a.stats.ResponseTimes) > 1000 {
		a.stats.ResponseTimes = a.stats.ResponseTimes[1:]
	}
	a.stats.LastActivity = time.Now()
	a.statsMu.Unlock()
}

func (a *tunnelAgent) getOrCreateServiceStats(name string) *ServiceStats {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	if s, ok := a.stats.ServiceStats[name]; ok {
		return s
	}
	s := &ServiceStats{}
	a.stats.ServiceStats[name] = s
	return s
}

func (a *tunnelAgent) snapshotStats() *Stats {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()

	snapshot := &Stats{
		Connections:  atomic.LoadInt64(&a.stats.Connections),
		Streams:      atomic.LoadInt64(&a.stats.Streams),
		BytesIn:      atomic.LoadInt64(&a.stats.BytesIn),
		BytesOut:     atomic.LoadInt64(&a.stats.BytesOut),
		Requests:     atomic.LoadInt64(&a.stats.Requests),
		Errors:       atomic.LoadInt64(&a.stats.Errors),
		LastActivity: a.stats.LastActivity,
	}

	if len(a.stats.ResponseTimes) > 0 {
		snapshot.ResponseTimes = append([]time.Duration(nil), a.stats.ResponseTimes...)
	}

	if len(a.stats.ServiceStats) > 0 {
		snapshot.ServiceStats = make(map[string]*ServiceStats, len(a.stats.ServiceStats))
		for name, s := range a.stats.ServiceStats {
			snapshot.ServiceStats[name] = &ServiceStats{
				Requests:    atomic.LoadInt64(&s.Requests),
				Errors:      atomic.LoadInt64(&s.Errors),
				BytesIn:     atomic.LoadInt64(&s.BytesIn),
				BytesOut:    atomic.LoadInt64(&s.BytesOut),
				LastRequest: s.LastRequest,
				ResponseTimes: func() []time.Duration {
					if len(s.ResponseTimes) == 0 {
						return nil
					}
					return append([]time.Duration(nil), s.ResponseTimes...)
				}(),
			}
		}
	}

	return snapshot
}
