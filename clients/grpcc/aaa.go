package grpcc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pubgo/lava"
)

// Interface grpc client interface
type Interface interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
	Middleware(mm ...lava.Middleware)
}
