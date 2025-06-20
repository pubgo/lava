package grpcc

import (
	"fmt"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
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

func createConn(cfg *grpccconfig.Cfg, log log.Logger, mm []lava.Middleware) (r result.Result[grpc.ClientConnInterface]) {
	addr := buildTarget(cfg.Service)

	var logMsg = func(e *zerolog.Event) {
		e.Any(logkey.Service, cfg.Service)
		e.Any("config", cfg.Client)
		e.Str("addr", addr)
	}

	defer func() {
		if r.IsOK() {
			log.Info().
				Func(logMsg).Msg("succeed to create grpc client")
		} else {
			log.Err(r.GetErr()).
				Func(logMsg).Msg("failed to create grpc client")
		}
	}()

	opts := cfg.Client.ToOpts()
	opts = append(opts, grpc.WithResolvers(cfg.Resolvers...))
	opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptor(mm)))
	opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptor(mm)))

	conn := result.Wrap(grpc.NewClient(addr, opts...)).
		MapErr(func(err error) error {
			return errors.Wrapf(err, "failed to dial grpc server, target=%s", addr)
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	return r.WithValue(conn)
}
