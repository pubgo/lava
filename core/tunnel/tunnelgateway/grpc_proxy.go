package tunnelgateway

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

const (
	// grpcRoutePrefix 保留别名，实际逻辑见 tunnel.GRPCRoutePrefix。
	grpcRoutePrefix = tunnel.GRPCRoutePrefix
)

// startGRPCProxy 在 GRPCPort 上监听 TCP 连接，按路由前缀转发到对应 Agent 的 gRPC 端点。
func (g *tunnelGateway) startGRPCProxy() {
	defer g.wg.Done()

	addr := fmt.Sprintf(":%d", g.cfg.GRPCPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error().Err(err).Str("addr", addr).Msg("GRPC proxy listen failed")
		return
	}
	g.grpcListener = ln

	log.Info().Str("addr", addr).Msg("GRPC proxy server started")

	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-g.stopCh:
				return
			default:
				log.Warn().Err(err).Msg("GRPC proxy accept failed")
				continue
			}
		}

		go g.handleGRPCConnection(conn)
	}
}

func (g *tunnelGateway) handleGRPCConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warn().Err(err).Msg("GRPC proxy: failed to close client connection")
		}
	}()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Warn().Err(err).Msg("GRPC proxy: failed to set read deadline")
		return
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Warn().Err(err).Msg("GRPC proxy: failed to read routing line")
		return
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		log.Warn().Err(err).Msg("GRPC proxy: failed to clear read deadline")
		return
	}

	line = strings.TrimSpace(line)
	serviceName, ok := tunnel.ParseGRPCRouteLine(line)
	if !ok {
		log.Warn().Str("line", line).Msg("GRPC proxy: invalid routing line, expected TUNNEL <service>")
		return
	}

	token := ""
	if g.authProvider != nil {
		authLine, err := reader.ReadString('\n')
		if err != nil {
			log.Warn().Err(err).Str("service", serviceName).Msg("GRPC proxy: failed to read auth line")
			return
		}
		var authOK bool
		token, authOK = tunnel.ParseGRPCAuthLine(authLine)
		if !authOK {
			log.Warn().Str("service", serviceName).Msg("GRPC proxy: missing or invalid auth line")
			return
		}
	}

	if err := g.authorizeClient(serviceName, token); err != nil {
		log.Warn().Err(err).Str("service", serviceName).Msg("GRPC proxy: unauthorized")
		return
	}

	if !g.rateLimiter.Allow(serviceName) {
		log.Warn().Str("service", serviceName).Msg("GRPC proxy: rate limit exceeded")
		return
	}

	wrapped := &bufferedConn{Conn: conn, r: reader}
	if err := g.Forward(context.Background(), serviceName, tunnel.EndpointTypeGRPC, wrapped); err != nil {
		log.Warn().Err(err).Str("service", serviceName).Msg("GRPC proxy: forward failed")
	}
}

// bufferedConn 在读取完路由行后，继续从 bufio.Reader 中读取剩余数据。
type bufferedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	if c.r.Buffered() > 0 {
		return c.r.Read(p)
	}
	return c.Conn.Read(p)
}

// 确保 bufferedConn 满足 io.Reader 语义（Forward 使用 io.Copy）。
var _ io.Reader = (*bufferedConn)(nil)
