package gatewayserver

import "testing"

func TestDefaultCfgGRPCPassthrough(t *testing.T) {
	t.Parallel()
	cfg := defaultCfg()
	if !cfg.GRPCPassthrough {
		t.Fatal("expected grpc_passthrough default true")
	}
}
