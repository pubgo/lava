package tunnelgateway

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestAuthorizeClient(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)

	if err := g.authorizeClient("svc", ""); err != nil {
		t.Fatalf("nil provider should allow empty token: %v", err)
	}

	g.authProvider = &fakeAuth{}
	if err := g.authorizeClient("svc", ""); !errors.Is(err, tunnel.ErrAuthFailed) {
		t.Fatalf("empty token err=%v", err)
	}

	g.authProvider = &fakeAuth{validateErr: tunnel.ErrAuthFailed}
	if err := g.authorizeClient("svc", "bad"); !errors.Is(err, tunnel.ErrAuthFailed) {
		t.Fatalf("validate err=%v", err)
	}

	g.authProvider = &fakeAuth{authorizeErr: tunnel.ErrAuthFailed}
	if err := g.authorizeClient("svc", "tok"); !errors.Is(err, tunnel.ErrAuthFailed) {
		t.Fatalf("authorize err=%v", err)
	}

	g.authProvider = &fakeAuth{}
	if err := g.authorizeClient("svc", "tok"); err != nil {
		t.Fatalf("valid token: %v", err)
	}
}

func TestWriteAuthError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeAuthError(rr)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "unauthorized") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
