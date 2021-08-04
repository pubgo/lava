package grpcc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type DialOptions = []grpc.DialOption

type Client interface {
	Check(opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error)
	Get() (*grpc.ClientConn, error)
	Close() error
}
