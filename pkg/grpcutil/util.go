package grpcutil

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

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
