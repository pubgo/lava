package zrpcs

import (
	"testing"

	"github.com/pubgo/funk/v2/log"
)

func TestDefaultCfg(t *testing.T) {
	cfg := defaultCfg()
	if cfg.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("url=%q", cfg.URL)
	}
}

func TestServiceStringAndInit(t *testing.T) {
	s := &serviceImpl{}
	s.init(Params{
		Conf: &Config{URL: "nats://example:4222"},
		Log:  log.GetLogger("zrpcs-test"),
	})
	if s.String() != "zrpc-server" {
		t.Fatalf("name=%q", s.String())
	}
	if s.conf.URL != "nats://example:4222" {
		t.Fatalf("conf url=%q", s.conf.URL)
	}
	if len(s.mw) < 4 {
		t.Fatalf("expected default middlewares, got %d", len(s.mw))
	}
}

func TestStart_InvalidURL(t *testing.T) {
	s := &serviceImpl{}
	s.init(Params{
		Conf: &Config{URL: "://bad"},
		Log:  log.GetLogger("zrpcs-test"),
	})
	if err := s.start(); err == nil {
		t.Fatal("expected connect error")
	}
}
