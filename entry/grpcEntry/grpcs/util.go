package grpcs

import (
	"fmt"

	"google.golang.org/grpc"
	srvChannel "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/profiling"
	srvProfile "google.golang.org/grpc/profiling/service"
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
	srvChannel.RegisterChannelzServiceToServer(s)
	profiling.Enable(true)
	srvProfile.Init(&srvProfile.ProfilingConfig{})

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
