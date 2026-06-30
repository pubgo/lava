package tunnel

import "testing"

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

func TestTokenAuthProviderEmpty(t *testing.T) {
	if _, err := NewTokenAuthProvider(); err == nil {
		t.Fatalf("expected error when no tokens provided")
	}
}
