package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
)

type Option func(cli *Client)

func WithDial(fn func(srv string, cfg grpcc_config.Cfg) (grpc.ClientConnInterface, error)) func(cli *Client) {
	return func(cli *Client) { cli.dial = fn }
}
