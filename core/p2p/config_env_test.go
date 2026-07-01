package p2p_test

import (
	"os"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("P2P_STUN_URLS", "stun:a:3478,stun:b:3478")
	t.Setenv("P2P_TURN_URL", "turn:example:3478")
	t.Setenv("P2P_TURN_USER", "u")
	t.Setenv("P2P_TURN_PASS", "p")
	t.Setenv("P2P_ICE_TIMEOUT", "12s")
	t.Setenv("P2P_INSECURE", "true")
	t.Setenv("P2P_AUTH_TOKEN", "tok")

	cfg := p2p.ConfigFromEnv()
	if len(cfg.STUNURLs) != 2 || cfg.STUNURLs[0] != "stun:a:3478" {
		t.Fatalf("stun urls: %v", cfg.STUNURLs)
	}
	if cfg.TURN.URL != "turn:example:3478" || cfg.TURN.Username != "u" {
		t.Fatalf("turn: %+v", cfg.TURN)
	}
	if cfg.ICETimeout != 12*time.Second {
		t.Fatalf("timeout: %v", cfg.ICETimeout)
	}
	if !cfg.Insecure || cfg.AuthToken != "tok" {
		t.Fatalf("insecure=%v token=%q", cfg.Insecure, cfg.AuthToken)
	}
}

func TestConfigFromEnvTLS(t *testing.T) {
	t.Setenv("P2P_CERT_FILE", "/etc/p2p/cert.pem")
	t.Setenv("P2P_KEY_FILE", "/etc/p2p/key.pem")
	cfg := p2p.ConfigFromEnv()
	opts := cfg.TransportOptions()
	if !opts.EnableTLS || opts.CertFile != "/etc/p2p/cert.pem" {
		t.Fatalf("tls opts: %+v", opts)
	}
}

func TestConfigFromEnvTURNDisabled(t *testing.T) {
	t.Setenv("P2P_TURN_DISABLED", "true")
	cfg := p2p.ConfigFromEnv()
	if !cfg.TURN.Disabled {
		t.Fatalf("turn disabled: %+v", cfg.TURN)
	}
	if len(cfg.ICEURLs()) != len(cfg.STUNURLs) {
		t.Fatalf("ICEURLs should exclude TURN: %v", cfg.ICEURLs())
	}
}

func TestAuthTokenFromEnvFallback(t *testing.T) {
	_ = os.Unsetenv("P2P_AUTH_TOKEN")
	t.Setenv("TUNNEL_AUTH_TOKEN", "tunnel-tok")
	if got := p2p.AuthTokenFromEnv(); got != "tunnel-tok" {
		t.Fatalf("got %q", got)
	}
}
