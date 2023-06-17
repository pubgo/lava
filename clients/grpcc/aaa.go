package grpcc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
)

const Name = "grpcc"

type Config = grpcc_config.Cfg

// Client grpc client interface
type Client interface {
	grpc.ClientConnInterface
	Healthy(ctx context.Context) error
}
