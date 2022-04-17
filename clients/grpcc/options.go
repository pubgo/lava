package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
)

type Option func(cli *Client)

func WithDirect() func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.Scheme = grpcc_resolver.DirectScheme }
}

func WithDns() func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.Scheme = grpcc_resolver.DnsScheme }
}

func WithK8s() func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.Scheme = grpcc_resolver.K8sScheme }
}

func WithDiscov() func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.Scheme = grpcc_resolver.DiscovScheme }
}

func WithDial(fn func(srv string, cfg grpcc_config.Cfg) (grpc.ClientConnInterface, error)) func(cli *Client) {
	return func(cli *Client) { cli.dial = fn }
}
