package otel

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/recovery"
	ot_bridge "go.opentelemetry.io/otel/bridge/opentracing"
)

func init() {
	defer recovery.Exit()

	bridgeTracer, _ := ot_bridge.NewTracerPair(nil)
	bridgeTracer.SetWarningHandler(func(msg string) {})
	opentracing.SetGlobalTracer(bridgeTracer)
}
