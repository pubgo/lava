package httputil

import "testing"

func TestDefaultCfgBodyLimit(t *testing.T) {
	t.Parallel()
	cfg := DefaultCfg()
	if cfg.Http.BodyLimit != DefaultBodyLimit {
		t.Fatalf("BodyLimit = %d, want %d", cfg.Http.BodyLimit, DefaultBodyLimit)
	}
}
