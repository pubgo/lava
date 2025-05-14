package grpcc

import (
	"fmt"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/clients/grpcc/grpccconfig"
	"github.com/pubgo/lava/clients/grpcc/grpccresolver"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/lava"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func buildTarget(cfg *grpccconfig.ServiceCfg) string {
	addr := cfg.Addr
	scheme := grpccresolver.DirectScheme
	if cfg.Scheme != "" {
		scheme = cfg.Scheme
	}

	switch scheme {
	case grpccresolver.DiscoveryScheme:
		return grpccresolver.BuildDiscoveryTarget(addr)
	case grpccresolver.DirectScheme:
		return grpccresolver.BuildDirectTarget(cfg.Name, addr)
	case grpccresolver.K8sScheme, grpccresolver.DnsScheme:
		return fmt.Sprintf("dns:///%s", addr)
	default:
		return addr
	}
}

func createConn(cfg *grpccconfig.Cfg, log log.Logger, mm []lava.Middleware) (_ grpc.ClientConnInterface, gErr error) {
	addr := buildTarget(cfg.Service)

	var logMsg = func(e *zerolog.Event) {
		e.Any(logkey.Service, cfg.Service)
		e.Any("config", cfg.Client)
		e.Str("addr", addr)
	}

	defer func() {
		if gErr == nil {
			log.Info().
				Func(logMsg).Msg("succeed to create grpc client")
		} else {
			log.Err(gErr).
				Func(logMsg).Msg("failed to create grpc client")
		}
	}()

	opts := append(
		cfg.Client.ToOpts(),
		grpc.WithResolvers(cfg.Resolvers...),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(mm)),
		grpc.WithChainStreamInterceptor(streamInterceptor(mm)),
	)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc dial failed, target=>%s", addr)
	}

	return conn, nil
}
