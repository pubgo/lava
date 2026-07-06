//go:build legacy
// +build legacy

package gatewayserver

import (
	"testing"

	"github.com/pubgo/funk/v2/log"
)

func TestResolveConfigLegacyGrpcServer(t *testing.T) {
	t.Parallel()

	cfg := ResolveConfig(CombinedConfigLoader{
		GrpcServer: &Config{EnablePrintRouter: true},
	}, log.GetLogger("test"))
	if !cfg.EnablePrintRouter {
		t.Fatal("expected legacy grpc_server config to apply")
	}
}
