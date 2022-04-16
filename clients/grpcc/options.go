package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
)

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

func WithRegistry(name string) func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.registry = name }
}

func WithClientType(typ interface{}) func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.clientType = typ }
}

func WithNewClientFunc(fn func(cc grpc.ClientConnInterface) interface{}) func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.newClient = fn }
}

func WithBeforeDial(fn func()) func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.beforeDial = fn }
}

func WithAfterDial(fn func()) func(cfg *grpcc_config.Cfg) {
	return func(cfg *grpcc_config.Cfg) { cfg.afterDial = fn }
}

func WithDial(fn func(addr string, cfg *grpcc_config.Cfg, plugins ...string) (*grpc.ClientConn, error)) func(cli *Client) {
	return func(cli *Client) { cli.dial = fn }
}
