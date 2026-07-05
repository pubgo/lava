package grpcservercmd

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/core/supervisor/bootstrap"
	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/pubgo/lava/v2/servers/gatewayserver"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "grpc",
		Short: cliutil.UsageDesc("gateway server %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			di.Provide(gatewayserver.New)
			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service
			}))

			return bootstrap.RunServices(ctx, params.LC, params.Services)
		},
	}
}
