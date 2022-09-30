package gidsrv

import (
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/logging"
)

func NewClient(cfg *grpcc_config.Cfg, log *logging.Logger) *Client {
	return &Client{gidpb.NewIdClient(grpcc.New(cfg, log))}
}

type Client struct {
	gidpb.IdClient
}
