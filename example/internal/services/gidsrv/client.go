package gidsrv

import (
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/service"
)

func NewClient(cfg *grpcc_config.Cfg, log *logging.Logger, middlewares map[string]service.Middleware) *Client {
	return &Client{gidpb.NewIdClient(grpcc.New(cfg, log, middlewares))}
}

type Client struct {
	gidpb.IdClient
}
