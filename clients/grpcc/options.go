package grpcc

import "github.com/pubgo/lava/clients/grpcc/resolver"

func WithDirect() func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.buildScheme = resolver.DirectScheme }
}

func WithDns() func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.buildScheme = resolver.DnsScheme }
}

func WithK8s() func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.buildScheme = resolver.K8sScheme }
}

func WithDiscov() func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.buildScheme = resolver.DiscovScheme }
}

func WithRegistry(name string) func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.registry = name }
}

func WithClientType(typ interface{}) func(cfg *Cfg) {
	return func(cfg *Cfg) { cfg.clientType = typ }
}
