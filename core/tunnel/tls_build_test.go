package tunnel_test

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestBuildClientTLSConfig_MinVersion(t *testing.T) {
	cfg := tunnel.BuildClientTLSConfig(&tunnel.TransportOptions{
		MinVersion: "TLS13",
		Insecure:   true,
	})
	if cfg.MinVersion != 0x0304 { // tls.VersionTLS13
		t.Fatalf("min version=%x", cfg.MinVersion)
	}
}

func TestMergeTransportOptions_TLSAdvanced(t *testing.T) {
	opts := tunnel.MergeTransportOptions(nil, tunnel.TLSConfig{
		MinVersion:       "TLS12",
		CipherSuites:     []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		ClientAuth:       "RequireAndVerifyClientCert",
		SessionCacheSize: 128,
	})
	if opts.MinVersion != "TLS12" {
		t.Fatalf("min=%q", opts.MinVersion)
	}
	if len(opts.CipherSuites) != 1 {
		t.Fatalf("ciphers=%v", opts.CipherSuites)
	}
	if opts.ClientAuth != "RequireAndVerifyClientCert" {
		t.Fatalf("client auth=%q", opts.ClientAuth)
	}
	if opts.SessionCacheSize != 128 {
		t.Fatalf("cache=%d", opts.SessionCacheSize)
	}
}
