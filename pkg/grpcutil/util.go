package grpcutil

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pubgo/funk/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	srvProfile "google.golang.org/grpc/profiling/service"
	"google.golang.org/grpc/reflection"
)

const (
	ClientNameKey     = "client-name"
	ClientIpKey       = "client-ip"
	ClientHostnameKey = "client-hostname"
	ClientPathKey     = "client-path"
	ClientVersionKey  = "client-version"
)

// WithClientApp 获取对端应用名称
func WithClientApp(ctx context.Context, name string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ClientNameKey, name)
}

func WithClientIp(ctx context.Context, ip string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ClientIpKey, ip)
}

// ClientName 获取对端应用名称
func ClientName(md metadata.MD) string {
	return HeaderGet(md, ClientNameKey)
}

// ClientIP 获取对端ip
func ClientIP(md metadata.MD) string {
	return HeaderGet(md, ClientIpKey)
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
	assert.Must(srvProfile.Init(&srvProfile.ProfilingConfig{Enabled: true, Server: s}))
}

func EnableAdmin(s grpc.ServiceRegistrar) (cleanup func(), _ error) {
	return admin.Register(s)
}

// IsGRPCRequest returns true if the message is considered to be a GRPC message
func IsGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), DefaultContentType)
}

// ListGRPCResources is a helper function that lists all URLs that are registered on gRPC server.
//
// This makes it easy to register all the relevant routes in your HTTP router of choice.
func ListGRPCResources(server *grpc.Server) []string {
	var ret []string
	for serviceName, serviceInfo := range server.GetServiceInfo() {
		for _, methodInfo := range serviceInfo.Methods {
			ret = append(ret, fmt.Sprintf("/%s/%s", serviceName, methodInfo.Name))
		}
	}
	return ret
}
