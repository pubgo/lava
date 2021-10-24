package grpcc

import (
	"strings"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/pkg/ctxutil"
)

// serviceFromMethod returns the service
// /service.Foo/Bar => service
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	if m[0] != '/' {
		return m
	}

	parts := strings.Split(m, "/")
	if len(parts) < 3 {
		return m
	}

	parts = strings.Split(parts[1], ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func HealthCheck(srv string, conn *grpc.ClientConn) error {
	_, err := grpc_health_v1.NewHealthClient(conn).Check(ctxutil.Timeout(), &grpc_health_v1.HealthCheckRequest{Service: srv})
	return xerror.Wrap(err)
}
