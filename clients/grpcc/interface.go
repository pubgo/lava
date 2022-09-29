package grpcc

import (
	"context"

	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"
)

// Interface grpc client interface
type Interface interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
	Middleware(mm ...service.Middleware)
}
