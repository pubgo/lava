package jaeger

import (
	"github.com/pubgo/lug/tracing"

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
	cfg.ServiceName = "service_name"

	tracer := xerror.PanicErr(New(cfg)).(tracing.Tracer)

	span := tracer.RootSpan("start")
	span.LogKV("hello", "world")
	span.LogKV("error", "world")
	span.LogKV("error111", "world")
	span.SetTag("error", "错了")
	span.SetTag("error111", "错了")
	span.Finish()

	span1 := span.CreateFollows("Follows", tracing.Tags{"name": "Follows"})
	span1.LogKV("hello", "world")
	span1.LogKV("error", "world")
	span1.SetTag("error", "错了")
	span1.LogKV("key", "value")
	span1.SetTag("http", "get")
	span1.Finish()

	span2 := span1.CreateChild("Child", tracing.Tags{"name": "Follows"})
	span2.LogKV("hello", "world")
	span2.LogKV("error", "world")
	span2.SetTag("error", "错了")
	span2.LogKV("key", "value")
	span2.SetTag("http", "get")
	span2.Finish()

	span3 := span.CreateChild("Child3", tracing.Tags{"name": "Follows"})
	span3.LogKV("hello", "world")
	span3.LogKV("error", "world")
	span3.SetTag("error", "错了")
	span3.LogKV("key", "value")
	span3.SetTag("http", "get")
	span3.Finish()
}
