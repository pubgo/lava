package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"

	"testing"
)

func TestLogger(t *testing.T) {
	xlog.Info("hello")
}

func TestName(t *testing.T) {
	defer xerror.RespDebug()

	var cfg = GetDefaultCfg()
	cfg.ServiceName = "test_span"

	tracer := xerror.PanicErr(New(cfg)).(opentracing.Tracer)

	span := tracer.StartSpan("start")
	defer span.Finish()

	span.LogKV("key", "value")
	span.SetTag("http", "get")


	span = tracer.StartSpan("start1")
	defer span.Finish()

	span.LogKV("key1", "value")
	span.SetTag("http", "get")
}
