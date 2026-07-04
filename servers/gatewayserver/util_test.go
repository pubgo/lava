package gatewayserver

import "testing"

func TestServiceFromMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		method  string
		service string
	}{
		{method: "/echo.Echo/Ping", service: "echo.Echo"},
		{method: "/pkg.Service/Method", service: "pkg.Service"},
		{method: "", service: ""},
	}
	for _, tt := range tests {
		if got := serviceFromMethod(tt.method); got != tt.service {
			t.Fatalf("serviceFromMethod(%q) = %q, want %q", tt.method, got, tt.service)
		}
	}
}
