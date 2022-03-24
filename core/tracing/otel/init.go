package otel

import (
	"github.com/opentracing/opentracing-go"
	ot_bridge "go.opentelemetry.io/otel/bridge/opentracing"
)

func init() {
	bridgeTracer, _ := ot_bridge.NewTracerPair(nil)
	bridgeTracer.SetWarningHandler(func(msg string) {})
	opentracing.SetGlobalTracer(bridgeTracer)
}
