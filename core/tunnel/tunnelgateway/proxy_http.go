package tunnelgateway

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// linkRewritePatterns 用于重写 HTML 响应中的链接
var linkRewritePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(href=["'])/debug/`),
	regexp.MustCompile(`(src=["'])/debug/`),
	regexp.MustCompile(`(action=["'])/debug/`),
	regexp.MustCompile(`(fetch\(["'])/debug/`),
	regexp.MustCompile(`(url:\s*["'])/debug/`),
	regexp.MustCompile(`(")/debug/`),
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

		if err := g.authorizeClient(serviceName, tunnel.ClientTokenFromRequest(r)); err != nil {
			writeAuthError(w)
			return
		}

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

// proxyToAgent 将 HTTP 请求代理到 Agent
func (g *tunnelGateway) proxyToAgent(w http.ResponseWriter, r *http.Request, svc *registeredService, endpointType tunnel.EndpointType, subPath string) {
	ctx := r.Context()
	serviceName := svc.info.Name

	isWebSocket := strings.EqualFold(r.Header.Get("Upgrade"), "websocket")

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

	priority := 5
	switch endpointType {
	case tunnel.EndpointTypeDebug:
		priority = 3
	case tunnel.EndpointTypeGRPC:
		priority = 4
	}

	stream, err := svc.session.OpenWithPriority(ctx, priority)
	if err != nil {
		stream, err = svc.session.Open(ctx)
		if err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: Failed to open stream to agent")
			http.Error(w, fmt.Sprintf("failed to open stream: %v", err), http.StatusInternalServerError)
			return
		}
	}

	log.Debug().Str("service", serviceName).Int("priority", priority).Msg("Gateway: Stream opened to agent")

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

	r.URL.Path = subPath
	r.RequestURI = subPath
	if r.URL.RawQuery != "" {
		r.RequestURI = subPath + "?" + r.URL.RawQuery
	}

	if isWebSocket {
		g.proxyWebSocket(w, r, stream, serviceName)
		return
	}

	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

	proxy := &httputil.ReverseProxy{
		Director:  func(req *http.Request) {},
		Transport: &streamRoundTripper{stream: stream, request: r},
		ModifyResponse: func(resp *http.Response) error {
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/json") {
				return nil
			}

			body, err := io.ReadAll(resp.Body)
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Str("service", serviceName).Msg("Gateway: failed to close response body")
			}
			if err != nil {
				return err
			}

			replacement := "${1}/" + serviceName + "/debug/"
			for _, pattern := range linkRewritePatterns {
				body = pattern.ReplaceAll(body, []byte(replacement))
			}

			resp.Body = io.NopCloser(bytes.NewReader(body))
			resp.ContentLength = int64(len(body))
			resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
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

// streamRoundTripper 实现 http.RoundTripper，通过 stream 转发请求
type streamRoundTripper struct {
	stream  tunnel.Stream
	request *http.Request
}

func (t *streamRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Write(t.stream); err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(t.stream), req)
}
