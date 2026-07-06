package gatewayserver

import (
	"testing"
)

func TestResolveConfigPrefersGatewayServer(t *testing.T) {
	t.Parallel()

	cfg := ResolveConfig(CombinedConfigLoader{
		GatewayServer: &Config{EnablePrintRouter: true},
		GrpcServer:    &Config{EnablePrintRouter: false},
	}, nil)
	if !cfg.EnablePrintRouter {
		t.Fatal("expected gateway_server to win")
	}
}

func TestResolveConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg := ResolveConfig(CombinedConfigLoader{}, nil)
	if !cfg.GRPCPassthrough {
		t.Fatal("expected default grpc_passthrough true")
	}
}

func TestDefaultCfgGRPCPassthrough(t *testing.T) {
	t.Parallel()
	cfg := defaultCfg()
	if !cfg.GRPCPassthrough {
		t.Fatal("expected grpc_passthrough default true")
	}
}
