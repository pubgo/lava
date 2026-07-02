package p2p_test

import (
	"strings"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
)

func TestTURNResolvedStatic(t *testing.T) {
	turn := p2p.TURNConfig{
		URL:      "turn:example:3478",
		Username: "u",
		Password: "p",
	}
	user, pass, err := turn.Resolved("peer-a")
	if err != nil || user != "u" || pass != "p" {
		t.Fatalf("static: user=%q pass=%q err=%v", user, pass, err)
	}
}

func TestTURNResolvedHMAC(t *testing.T) {
	turn := p2p.TURNConfig{
		URL:        "turn:example:3478",
		AuthSecret: "secret",
		CredTTL:    time.Hour,
	}
	user, pass, err := turn.Resolved("node-b")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(user, ":node-b") {
		t.Fatalf("user=%q", user)
	}
	if pass == "" {
		t.Fatal("empty password")
	}
	parts := strings.SplitN(user, ":", 2)
	if len(parts) != 2 {
		t.Fatalf("user=%q", user)
	}
}

func TestConfigFromEnvTURNSecret(t *testing.T) {
	t.Setenv("P2P_TURN_SECRET", "my-secret")
	t.Setenv("P2P_TURN_CRED_TTL", "2h")
	cfg := p2p.ConfigFromEnv()
	if cfg.TURN.AuthSecret != "my-secret" {
		t.Fatalf("secret=%q", cfg.TURN.AuthSecret)
	}
	if cfg.TURN.CredTTL != 2*time.Hour {
		t.Fatalf("ttl=%v", cfg.TURN.CredTTL)
	}
	user, pass, err := cfg.TURN.Resolved("peer-x")
	if err != nil || user == "" || pass == "" {
		t.Fatalf("resolved user=%q pass=%q err=%v", user, pass, err)
	}
}
