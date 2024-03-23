package gid_client

import (
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/resty"
	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
)

type Params struct {
	GrpcParam grpcc.Params
	Grpc      *GrpcConfig

	HttpParam resty.Params
	Http      *HttpConfig
}

type GrpcConfig struct {
	*grpcc.Config `yaml:",inline"`
}

type HttpConfig struct {
	*resty.Config `yaml:",inline"`
}

type Service struct {
	gidpb.IdClient
	resty.IClient
}

func New(p Params) *Service {
	cli := grpcc.New(p.Grpc.Config, p.GrpcParam)
	cli1 := resty.New(p.Http.Config, p.HttpParam)
	return &Service{
		IdClient: gidpb.NewIdClient(cli),
		IClient:  cli1,
	}
}
