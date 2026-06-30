package tunnel

import "testing"

func TestMergeTransportOptionsFromTLS(t *testing.T) {
	opts := MergeTransportOptions(nil, TLSConfig{
		Enabled:  true,
		CertFile: "/etc/certs/server.crt",
		KeyFile:  "/etc/certs/server.key",
		CAFile:   "/etc/certs/ca.crt",
	})

	if !opts.EnableTLS {
		t.Fatalf("expected EnableTLS true")
	}
	if opts.CertFile != "/etc/certs/server.crt" {
		t.Fatalf("unexpected cert file %q", opts.CertFile)
	}
	if opts.KeyFile != "/etc/certs/server.key" {
		t.Fatalf("unexpected key file %q", opts.KeyFile)
	}
	if opts.CAFile != "/etc/certs/ca.crt" {
		t.Fatalf("unexpected ca file %q", opts.CAFile)
	}
}

func TestMergeTransportOptionsExplicitWins(t *testing.T) {
	opts := MergeTransportOptions(&TransportOptions{
		EnableTLS: true,
		CertFile:  "/opts/cert.pem",
		KeyFile:   "/opts/key.pem",
	}, TLSConfig{
		CertFile: "/tls/cert.pem",
		KeyFile:  "/tls/key.pem",
	})

	if opts.CertFile != "/opts/cert.pem" {
		t.Fatalf("TransportOptions cert should take precedence, got %q", opts.CertFile)
	}
}

func TestGatewayConfigNormalize(t *testing.T) {
	cfg := &GatewayConfig{
		TLS: TLSConfig{Enabled: true, Insecure: true},
	}
	cfg.Normalize()
	if cfg.TransportOptions == nil || !cfg.TransportOptions.EnableTLS || !cfg.TransportOptions.Insecure {
		t.Fatalf("Normalize should merge TLS into TransportOptions")
	}
}

func TestApplyAuthTokenMetadata(t *testing.T) {
	md := ApplyAuthTokenMetadata(nil, "secret")
	if md["auth_token"] != "secret" {
		t.Fatalf("expected auth_token in metadata")
	}
}
