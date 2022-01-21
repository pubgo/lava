package sockets

import (
	"context"

	"google.golang.org/grpc"
)

type StreamHandler = grpc.StreamHandler
type UnaryHandler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)
