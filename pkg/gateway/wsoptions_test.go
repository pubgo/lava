package gateway

import "testing"

func TestWSOptionsFromConfig(t *testing.T) {
	opts := WSOptionsFromConfig(WSConfig{
		OriginPatterns:     []string{"example.com", "*.example.com"},
		InsecureSkipVerify: false,
	})
	if len(opts) != 2 {
		t.Fatalf("expected 2 options (subprotocols + origins), got %d", len(opts))
	}

	opts = WSOptionsFromConfig(WSConfig{InsecureSkipVerify: true})
	if len(opts) != 2 {
		t.Fatalf("expected 2 options (subprotocols + insecure), got %d", len(opts))
	}
}
