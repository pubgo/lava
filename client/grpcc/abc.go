package grpcc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Client interface {
	grpc_health_v1.HealthClient
	Get() (*grpc.ClientConn, error)
}
