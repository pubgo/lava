package tunnel

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// GRPCRoutePrefix 是 gateway gRPC 代理的路由前缀。
// 客户端须在 gRPC/HTTP2 流量前先发送一行：TUNNEL <service-name>\n
const GRPCRoutePrefix = "TUNNEL "

// GRPCRouteLine 返回写入 gateway gRPC 代理的首行路由指令。
func GRPCRouteLine(serviceName string) string {
	return GRPCRoutePrefix + serviceName + "\n"
}

// ParseGRPCRouteLine 解析客户端发送的路由行，返回 service 名。
func ParseGRPCRouteLine(line string) (serviceName string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, GRPCRoutePrefix) {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimPrefix(line, GRPCRoutePrefix))
	if name == "" {
		return "", false
	}
	return name, true
}

// GRPCAuthLine 返回 gRPC 代理的第二行鉴权指令（gateway 启用 auth 时必填）。
func GRPCAuthLine(token string) string {
	return GRPCAuthPrefix + token + "\n"
}

// ParseGRPCAuthLine 解析 gRPC 代理鉴权行，返回 token。
func ParseGRPCAuthLine(line string) (token string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, GRPCAuthPrefix) {
		return "", false
	}
	token = strings.TrimSpace(strings.TrimPrefix(line, GRPCAuthPrefix))
	return token, token != ""
}

// GRPCDialOptions 经 gateway gRPC 代理拨号的配置。
type GRPCDialOptions struct {
	GatewayAddr string
	// Token 非空时发送 Authorization 行；gateway 启用 auth 时必填。
	Token string
}

// GRPCContextDialer 返回可用于 grpc.WithContextDialer 的拨号函数。
//
// 拨号时 addr 参数即目标 service 名（配合 grpc passthrough target 使用）。
//
// 示例：
//
//	cc, err := grpc.NewClient("passthrough:///my-service",
//	    grpc.WithContextDialer(tunnel.GRPCContextDialer(tunnel.GRPCDialOptions{
//	        GatewayAddr: "gateway:9090",
//	        Token:       "secret",
//	    })),
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
func GRPCContextDialer(opts GRPCDialOptions) func(ctx context.Context, serviceName string) (net.Conn, error) {
	return func(ctx context.Context, serviceName string) (net.Conn, error) {
		if serviceName == "" {
			return nil, fmt.Errorf("tunnel: grpc service name required")
		}
		if opts.GatewayAddr == "" {
			return nil, fmt.Errorf("tunnel: grpc gateway addr required")
		}
		d := net.Dialer{}
		conn, err := d.DialContext(ctx, "tcp", opts.GatewayAddr)
		if err != nil {
			return nil, err
		}
		payload := GRPCRouteLine(serviceName)
		if opts.Token != "" {
			payload += GRPCAuthLine(opts.Token)
		}
		if _, err := conn.Write([]byte(payload)); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return conn, nil
	}
}
