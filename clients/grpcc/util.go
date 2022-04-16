package grpcc

import (
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/pkg/ctxutil"
)

func HealthCheck(srv string, conn grpc.ClientConnInterface) error {
	_, err := grpc_health_v1.NewHealthClient(conn).Check(ctxutil.Timeout(), &grpc_health_v1.HealthCheckRequest{Service: srv})
	return xerror.Wrap(err)
}
