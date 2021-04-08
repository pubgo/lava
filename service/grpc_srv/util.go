package golug_srv

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func EnableHealth(service string, s *grpc.Server) {
	healthCheck := health.NewServer()
	healthCheck.SetServingStatus(service, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthCheck)
}

func EnableReflection(s *grpc.Server) {
	reflection.Register(s)
}
