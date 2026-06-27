package tunnelgateway

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

func TestP2PRateLimitFromConfig(t *testing.T) {
	if got := p2pSignalRateLimit(nil); got != 60 {
		t.Fatalf("signal default=%d", got)
	}
	if got := p2pRegisterRateLimit(nil); got != 10 {
		t.Fatalf("register default=%d", got)
	}
	cfg := &tunnel.GatewayConfig{P2PSignalRateLimit: 30, P2PRegisterRateLimit: 2}
	if got := p2pSignalRateLimit(cfg); got != 30 {
		t.Fatalf("signal=%d", got)
	}
	if got := p2pRegisterRateLimit(cfg); got != 2 {
		t.Fatalf("register=%d", got)
	}
}

func TestP2PRegisterRateLimitEnforced(t *testing.T) {
	rl := p2pRegisterRateLimit(&tunnel.GatewayConfig{P2PRegisterRateLimit: 1})
	limiter := ratelimit.New(rl)
	if !limiter.Allow("agent-1") {
		t.Fatal("first register should pass")
	}
	if limiter.Allow("agent-1") {
		t.Fatal("second register should be limited")
	}
}
