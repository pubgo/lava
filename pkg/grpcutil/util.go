package grpcutil

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
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
	return HeaderGet(md, "client-app")
}

// GetClientIP 获取对端ip
func GetClientIP(md metadata.MD) string {
	return HeaderGet(md, "client-ip")
}

func EnableHealth(srv string, s *grpc.Server) {
	healthCheck := health.NewServer()
	healthCheck.SetServingStatus(srv, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthCheck)
}

func EnableReflection(s *grpc.Server) {
	reflection.Register(s)
}

func EnableDebug(s *grpc.Server) {
	grpc.EnableTracing = true
	service.RegisterChannelzServiceToServer(s)
}

func EnableAdmin(s grpc.ServiceRegistrar) (cleanup func(), _ error) {
	return admin.Register(s)
}

// IsGRPCRequest returns true if the message is considered to be
// a GRPC message
func IsGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}
