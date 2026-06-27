package ratelimit_test

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

func TestLimiterDefaultAllow(t *testing.T) {
	rl := ratelimit.New(2)
	if !rl.Allow("a") || !rl.Allow("a") {
		t.Fatal("expected first two allows")
	}
	if rl.Allow("a") {
		t.Fatal("expected third denied without refill")
	}
}

func TestLimiterPerKeyLimit(t *testing.T) {
	rl := ratelimit.New(100)
	rl.SetLimit("peer-a", 1)
	if !rl.Allow("peer-a") {
		t.Fatal("first allow")
	}
	if rl.Allow("peer-a") {
		t.Fatal("second should deny")
	}
	if !rl.Allow("peer-b") {
		t.Fatal("other key should use default")
	}
}
