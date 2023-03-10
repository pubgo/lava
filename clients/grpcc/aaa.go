package grpcc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pubgo/lava"
)

// Client grpc client interface
type Client interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
	Middleware(mm ...lava.Middleware)
}
