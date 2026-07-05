package bootstrap_test

import (
	"testing"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/bootstrap"
)

func TestRegisterAllIsIdempotent(t *testing.T) {
	t.Parallel()

	bootstrap.RegisterAll()
	before := len(debug.App().Stack()[0])
	bootstrap.RegisterAll()
	after := len(debug.App().Stack()[0])

	if before == 0 {
		t.Fatal("expected debug routes to be registered")
	}
	if after != before {
		t.Fatalf("RegisterAll registered routes twice: before=%d after=%d", before, after)
	}
}
