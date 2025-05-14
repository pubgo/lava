package grpcc

import (
	"context"

	"github.com/pubgo/lava/clients/grpcc/grpccconfig"
	"google.golang.org/grpc"
)

const Name = "grpcc"

type Config = grpccconfig.Cfg

// Client grpc client interface
type Client interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
}
