package tunnel

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

// BuildClientTLSConfig 根据 TransportOptions 构建客户端 tls.Config。
func BuildClientTLSConfig(opts *TransportOptions) *tls.Config {
	cfg := &tls.Config{InsecureSkipVerify: opts != nil && opts.Insecure}
	if opts == nil {
		return cfg
	}
	applyTLSVersionAndCiphers(cfg, opts)
	if opts.CAFile != "" && !opts.Insecure {
		caPEM, err := os.ReadFile(opts.CAFile)
		if err == nil {
			pool := x509.NewCertPool()
			if pool.AppendCertsFromPEM(caPEM) {
				cfg.RootCAs = pool
			}
		}
	}
	return cfg
}

// BuildServerTLSConfig 根据 TransportOptions 构建服务端 tls.Config。
func BuildServerTLSConfig(opts *TransportOptions) (*tls.Config, error) {
	if opts == nil {
		return nil, fmt.Errorf("tunnel: transport options required for server TLS")
	}
	cfg := &tls.Config{}
	if opts.CertFile != "" && opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("tunnel: load cert/key: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	applyTLSVersionAndCiphers(cfg, opts)
	if ca, err := parseClientAuth(opts.ClientAuth); err != nil {
		return nil, err
	} else if ca != tls.NoClientCert {
		cfg.ClientAuth = ca
		if opts.CAFile != "" {
			caPEM, err := os.ReadFile(opts.CAFile)
			if err != nil {
				return nil, fmt.Errorf("tunnel: read ca for client auth: %w", err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(caPEM) {
				return nil, fmt.Errorf("tunnel: invalid ca pem in %s", opts.CAFile)
			}
			cfg.ClientCAs = pool
		}
	}
	if opts.SessionCacheSize > 0 {
		cfg.ClientSessionCache = tls.NewLRUClientSessionCache(opts.SessionCacheSize)
	}
	return cfg, nil
}

func applyTLSVersionAndCiphers(cfg *tls.Config, opts *TransportOptions) {
	if v := parseTLSVersion(opts.MinVersion); v != 0 {
		cfg.MinVersion = v
	}
	if len(opts.CipherSuites) > 0 {
		var suites []uint16
		for _, name := range opts.CipherSuites {
			if id, ok := tlsCipherSuite(name); ok {
				suites = append(suites, id)
			}
		}
		if len(suites) > 0 {
			cfg.CipherSuites = suites
		}
	}
}

func parseTLSVersion(s string) uint16 {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TLS10", "TLS1.0":
		return tls.VersionTLS10
	case "TLS11", "TLS1.1":
		return tls.VersionTLS11
	case "TLS12", "TLS1.2":
		return tls.VersionTLS12
	case "TLS13", "TLS1.3":
		return tls.VersionTLS13
	default:
		return 0
	}
}

func parseClientAuth(mode string) (tls.ClientAuthType, error) {
	switch strings.TrimSpace(mode) {
	case "", "NoClientCert":
		return tls.NoClientCert, nil
	case "RequestClientCert":
		return tls.RequestClientCert, nil
	case "RequireAnyClientCert":
		return tls.RequireAnyClientCert, nil
	case "VerifyClientCertIfGiven":
		return tls.VerifyClientCertIfGiven, nil
	case "RequireAndVerifyClientCert":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return tls.NoClientCert, fmt.Errorf("tunnel: unknown client_auth %q", mode)
	}
}

func tlsCipherSuite(name string) (uint16, bool) {
	for _, cs := range tls.CipherSuites() {
		if cs.Name == name {
			return cs.ID, true
		}
	}
	for _, cs := range tls.InsecureCipherSuites() {
		if cs.Name == name {
			return cs.ID, true
		}
	}
	return 0, false
}
