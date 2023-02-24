package grpcc

import (
	"context"

	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
)

// Interface grpc client interface
type Interface interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
	Middleware(mm ...lava.Middleware)
}
