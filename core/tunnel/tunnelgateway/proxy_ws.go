package tunnelgateway

import (
	"io"
	"net/http"
	"sync"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// proxyWebSocket 处理 WebSocket 代理
func (g *tunnelGateway) proxyWebSocket(w http.ResponseWriter, r *http.Request, stream tunnel.Stream, serviceName string) {
	log.Debug().Str("service", serviceName).Str("path", r.URL.Path).Msg("Gateway: Starting WebSocket proxy")

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

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, clientConn); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: websocket copy client->agent failed")
		}
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("Gateway: failed to close stream")
		}
	}()

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
