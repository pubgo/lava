package grpcc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccconfig"
)

const Name = "grpcc"

type Config = grpccconfig.Cfg

// Client grpc client interface
type Client interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
}
