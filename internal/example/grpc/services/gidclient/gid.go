package gidclient

import (
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
)

type Config struct {
	grpcc.Config `yaml:",inline"`
}

type Service struct {
	gidpb.IdClient
}

func New(cfg *Config, p grpcc.Params) *Service {
	cli := grpcc.New(&cfg.Config, p)
	return &Service{IdClient: gidpb.NewIdClient(cli)}
}
