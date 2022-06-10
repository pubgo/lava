package grpcc

import (
	"github.com/pubgo/lava/internal/pkg/ctxutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func HealthCheck(srv string, conn grpc.ClientConnInterface) error {
	_, err := grpc_health_v1.NewHealthClient(conn).Check(ctxutil.Timeout(), &grpc_health_v1.HealthCheckRequest{Service: srv})
	return err
}
