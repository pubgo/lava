package grpcc

import "google.golang.org/grpc"

type GrpcClient interface {
	Ping() error
	Get() (*grpc.ClientConn, error)
}
