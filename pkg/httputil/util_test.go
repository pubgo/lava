package httputil

import (
	"testing"

	"github.com/pubgo/lava/v2/core/registry"
)

func TestDefaultCfgBodyLimit(t *testing.T) {
	t.Parallel()
	cfg := DefaultCfg()
	if cfg.Http.BodyLimit != registry.DefaultMaxMsgSize {
		t.Fatalf("BodyLimit = %d, want %d", cfg.Http.BodyLimit, registry.DefaultMaxMsgSize)
	}
}
