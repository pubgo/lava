package jump_test

import (
	"github.com/pubgo/lava/pkg/jump"
	"testing"
)

func TestHash(t *testing.T) {
	got := jump.Hash(123, 3)
	if got < 0 {
		t.Fatal("negative output")
	}
}

func TestHashString(t *testing.T) {
	got := jump.HashString("123", 3)
	if got < 0 {
		t.Fatal("negative output")
	}
}
