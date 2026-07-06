package gatewayserver

import "testing"

func TestResolveConfigPrefersGatewayServer(t *testing.T) {
	t.Parallel()

	cfg := ResolveConfig(ConfigLoader{
		GatewayServer: &Config{EnablePrintRouter: true},
	})
	if !cfg.EnablePrintRouter {
		t.Fatal("expected gateway_server config to apply")
	}
}

func TestResolveConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg := ResolveConfig(ConfigLoader{})
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
