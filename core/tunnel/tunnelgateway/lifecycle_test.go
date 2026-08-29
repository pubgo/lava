package tunnelgateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestServeHTTP_StatusJSON(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{
		ListenAddr: ":7007",
		Transport:  TransportYamux,
	}).(*tunnelGateway)
	g.status = tunnel.GatewayStatusRunning
	g.services["demo"] = &registeredService{
		info:  &tunnel.ServiceInfo{Name: "demo", ID: "id-1"},
		agent: "a1",
	}

	rr := httptest.NewRecorder()
	g.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/status", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != tunnel.GatewayStatusRunning.String() {
		t.Fatalf("status=%v", body["status"])
	}
	if int(body["num_services"].(float64)) != 1 {
		t.Fatalf("num_services=%v", body["num_services"])
	}
	if body["listen_addr"] != ":7007" {
		t.Fatalf("listen_addr=%v", body["listen_addr"])
	}
}
