package grpc_util

import (
	"context"

	"github.com/pubgo/lava/service"
	"google.golang.org/grpc/metadata"
)

// WithClientApp 获取对端应用名称
func WithClientApp(ctx context.Context, name string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "client-app", name)
}

func WithClientIp(ctx context.Context, ip string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "client-ip", ip)
}

// GetClientName 获取对端应用名称
func GetClientName(md metadata.MD) string {
	return service.HeaderGet(md, "client-app")
}

// GetClientIP 获取对端ip
func GetClientIP(md metadata.MD) string {
	return service.HeaderGet(md, "client-ip")
}
