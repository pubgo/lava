package grpcc

import (
	"context"

	"google.golang.org/grpc"
)

// Interface grpc client interface
type Interface interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
}
