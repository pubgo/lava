package gateway

import (
	"net/http/httptest"
	"testing"
)

func TestWSFrontend_matchRESTPath_POSTFallback(t *testing.T) {
	mux := NewMux()
	mux.routerTree.Add("POST", "/v1/greeter/hello", "/grpcweb.example.v1.GreeterService/SayHello", nil)
	mux.opts.handlers["/grpcweb.example.v1.GreeterService/SayHello"] = &methodWrapper{
		grpcFullMethod: "/grpcweb.example.v1.GreeterService/SayHello",
	}

	f := &wsFrontend{mux: mux}
	req := httptest.NewRequest("GET", "http://localhost/v1/greeter/hello", nil)

	match, values, ok := f.matchRESTPath(req)
	if !ok {
		t.Fatal("expected REST path match for GET handshake on POST route")
	}
	if match.Operation != "/grpcweb.example.v1.GreeterService/SayHello" {
		t.Fatalf("unexpected operation: %s", match.Operation)
	}
	if values == nil {
		t.Fatal("expected non-nil values")
	}
}

func TestWSFrontend_matchRESTPath_httpMethodOverride(t *testing.T) {
	mux := NewMux()
	mux.routerTree.Add("PUT", "/v1/items/{id}", "/example.v1.Service/Update", nil)
	mux.opts.handlers["/example.v1.Service/Update"] = &methodWrapper{
		grpcFullMethod: "/example.v1.Service/Update",
	}

	f := &wsFrontend{mux: mux}
	req := httptest.NewRequest("GET", "http://localhost/v1/items/42?http_method=PUT", nil)

	match, values, ok := f.matchRESTPath(req)
	if !ok {
		t.Fatal("expected match with http_method override")
	}
	if match.Operation != "/example.v1.Service/Update" {
		t.Fatalf("unexpected operation: %s", match.Operation)
	}
	if got := values.Get("id"); got != "42" {
		t.Fatalf("expected path var id=42, got %q", got)
	}
}
