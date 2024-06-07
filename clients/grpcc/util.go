package grpcc

import (
	"fmt"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
)

func buildTarget(cfg *grpcc_config.ServiceCfg) string {
	addr := cfg.Addr
	scheme := grpcc_resolver.DirectScheme
	if cfg.Scheme != "" {
		scheme = cfg.Scheme
	}

	switch scheme {
	case grpcc_resolver.DiscoveryScheme:
		return grpcc_resolver.BuildDiscoveryTarget(addr)
	case grpcc_resolver.DirectScheme:
		return grpcc_resolver.BuildDirectTarget(cfg.Name, addr)
	case grpcc_resolver.K8sScheme, grpcc_resolver.DnsScheme:
		return fmt.Sprintf("dns:///%s", addr)
	default:
		return addr
	}
}

func createConn(cfg *grpcc_config.Cfg, log log.Logger, mm []lava.Middleware) (grpc.ClientConnInterface, error) {
	addr := buildTarget(cfg.Service)

	ee := log.Info().
		Str(logkey.Service, cfg.Service.Name).
		Str("addr", addr)
	ee.Msg("grpc client init")

	conn, err := grpc.NewClient(addr, append(
		append(cfg.Client.ToOpts(), grpc.WithResolvers(cfg.Resolvers...)),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(mm)),
		grpc.WithChainStreamInterceptor(streamInterceptor(mm)))...)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc dial failed, target=>%s", addr)
	}

	ee.Msg("grpc client init ok")
	return conn, nil
}
