package tunnelagent

import (
	"io"
	"net"
	"sync"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// findEndpoint 根据 RequestMeta 中的 ServiceID 与端点类型查找本地端点。
// 当 ServiceID 非空时仅匹配对应服务；为空时保持兼容，返回第一个匹配类型的端点。
func findEndpoint(services map[string]*tunnel.ServiceInfo, meta tunnel.RequestMeta, want tunnel.EndpointType) *tunnel.Endpoint {
	et := want
	if meta.EndpointType != "" {
		et = meta.EndpointType
	}

	matchService := func(svc *tunnel.ServiceInfo) bool {
		if meta.ServiceID == "" {
			return true
		}
		return svc.ID == meta.ServiceID || svc.Name == meta.ServiceID
	}

	if meta.ServiceID != "" {
		for _, svc := range services {
			if !matchService(svc) {
				continue
			}
			for i := range svc.Endpoints {
				if svc.Endpoints[i].Type == et {
					return &svc.Endpoints[i]
				}
			}
		}
		return nil
	}

	for _, svc := range services {
		for i := range svc.Endpoints {
			if svc.Endpoints[i].Type == et {
				return &svc.Endpoints[i]
			}
		}
	}
	return nil
}

// normalizeAddress 将 ":port" 形式补全为 "127.0.0.1:port"。
func normalizeAddress(address string) string {
	if len(address) > 0 && address[0] == ':' {
		return "127.0.0.1" + address
	}
	return address
}

// proxyStreamToTCP 在 stream 与本地 TCP 服务之间双向转发。
func proxyStreamToTCP(stream tunnel.Stream, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("Failed to connect to local service")
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warn().Err(err).Str("address", address).Msg("Failed to close local connection")
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(conn, stream); err != nil {
			log.Warn().Err(err).Str("address", address).Msg("Failed to copy stream to local service")
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.CloseWrite(); err != nil {
				log.Warn().Err(err).Str("address", address).Msg("Failed to close write side")
			}
		}
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, conn); err != nil {
			log.Warn().Err(err).Str("address", address).Msg("Failed to copy local service to stream")
		}
	}()

	wg.Wait()
}
