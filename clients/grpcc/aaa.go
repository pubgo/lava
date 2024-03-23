package grpcc

import (
	"context"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"google.golang.org/grpc"
)

const Name = "grpcc"

type Config = grpcc_config.Cfg

// Client grpc client interface
type Client interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
}
