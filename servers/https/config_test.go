package https

import (
	"testing"

	"github.com/pubgo/lava/v2/core/registry"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

func TestDefaultHTTPConfigBodyLimit(t *testing.T) {
	t.Parallel()
	cfg := httputil.DefaultCfg()
	if cfg.Http.BodyLimit != registry.DefaultMaxMsgSize {
		t.Fatalf("BodyLimit = %d, want %d", cfg.Http.BodyLimit, registry.DefaultMaxMsgSize)
	}
}
