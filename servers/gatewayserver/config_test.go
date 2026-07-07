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
	if cfg.EnablePrintRouter {
		t.Fatal("expected enable_print_router default false")
	}
}
