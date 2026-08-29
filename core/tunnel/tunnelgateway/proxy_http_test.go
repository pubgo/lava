package tunnelgateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestCreateProxyHandler_Routing(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.services["demo"] = &registeredService{
		info:    &tunnel.ServiceInfo{Name: "demo", ID: "id-1"},
		session: &fakeSession{closed: true},
		agent:   "agent-1",
	}
	g.services["live"] = &registeredService{
		info:    &tunnel.ServiceInfo{Name: "live", ID: "id-2"},
		session: &fakeSession{},
		agent:   "agent-2",
	}

	h := g.createProxyHandler(tunnel.EndpointTypeHTTP)

	t.Run("service_list", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if int(body["count"].(float64)) != 2 {
			t.Fatalf("count=%v", body["count"])
		}
	})

	t.Run("missing_service", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/missing/x", nil))
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status=%d", rr.Code)
		}
	})

	t.Run("closed_session", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/demo/hello", nil))
		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("peer_list", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/p2p/peers", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), `"peers"`) {
			t.Fatalf("body=%s", rr.Body.String())
		}
	})

	t.Run("auth_required", func(t *testing.T) {
		g.authProvider = &fakeAuth{}
		defer func() { g.authProvider = nil }()
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestLinkRewritePatterns(t *testing.T) {
	html := `<a href="/debug/pprof">x</a><script src="/debug/js"></script>` +
		`<form action="/debug/x"></form>` +
		`fetch("/debug/api")` +
		`url: "/debug/y"` +
		`"/debug/z"`
	out := html
	for _, pattern := range linkRewritePatterns {
		out = pattern.ReplaceAllString(out, `${1}/svc/debug/`)
	}
	want := []string{
		`href="/svc/debug/pprof"`,
		`src="/svc/debug/js"`,
		`action="/svc/debug/x"`,
		`fetch("/svc/debug/api")`,
		`url: "/svc/debug/y"`,
		`"/svc/debug/z"`,
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in %q", w, out)
		}
	}
}
