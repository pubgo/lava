package tunnel

import (
	"net/http"
	"testing"
)

func TestTokenAuthProvider(t *testing.T) {
	p, err := NewTokenAuthProvider("secret-token")
	if err != nil {
		t.Fatalf("NewTokenAuthProvider: %v", err)
	}

	svc := &ServiceInfo{
		Name:     "test-svc",
		Metadata: map[string]string{"auth_token": "secret-token"},
	}
	if err := p.Authenticate(svc); err != nil {
		t.Fatalf("Authenticate should succeed: %v", err)
	}

	svc.Metadata["auth_token"] = "wrong"
	if err := p.Authenticate(svc); err == nil {
		t.Fatalf("expected auth failure for wrong token")
	}
}

func TestTokenAuthProviderAuthorize(t *testing.T) {
	p, err := NewTokenAuthProvider("secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Authorize("any-svc", "secret-token"); err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if err := p.Authorize("any-svc", "wrong"); err == nil {
		t.Fatal("expected authorize failure")
	}
}

func TestClientTokenFromRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example/", nil)
	req.Header.Set(HeaderTunnelToken, "via-header")
	if got := ClientTokenFromRequest(req); got != "via-header" {
		t.Fatalf("header token=%q", got)
	}
	req.Header.Del(HeaderTunnelToken)
	req.Header.Set("Authorization", "Bearer bearer-token")
	if got := ClientTokenFromRequest(req); got != "bearer-token" {
		t.Fatalf("bearer token=%q", got)
	}
}

func TestTokenAuthProviderEmpty(t *testing.T) {
	if _, err := NewTokenAuthProvider(); err == nil {
		t.Fatalf("expected error when no tokens provided")
	}
}

func TestGRPCAuthLine(t *testing.T) {
	line := GRPCAuthLine("tok")
	token, ok := ParseGRPCAuthLine(line)
	if !ok || token != "tok" {
		t.Fatalf("parse=%q ok=%v", token, ok)
	}
}
