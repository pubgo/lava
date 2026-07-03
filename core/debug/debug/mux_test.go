package debug

import "testing"

func TestIsLoopbackIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", false},
		{"192.168.1.1", false},
		{"", false},
		{"not-an-ip", false},
	}

	for _, tt := range tests {
		if got := isLoopbackIP(tt.ip); got != tt.want {
			t.Fatalf("isLoopbackIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestIsLoopbackIPRejectsPublicIP(t *testing.T) {
	t.Parallel()

	if isLoopbackIP("8.8.8.8") {
		t.Fatal("non-loopback IP must not bypass auth")
	}
}
