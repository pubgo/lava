package tracer

import (
	"context"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracingMaxContentLogSize = 512 * 1024 // Max log size for request and response body, especially for HTTP/RPC request.
	// defaultTextMapPropagator is the default propagator for context propagation between peers.
	defaultTextMapPropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

// MaxContentLogSize returns the max log size for request and response body, especially for HTTP/RPC request.
func MaxContentLogSize() int {
	return tracingMaxContentLogSize
}

// CheckSetDefaultTextMapPropagator sets the default TextMapPropagator if it is not set previously.
func CheckSetDefaultTextMapPropagator() {
	p := otel.GetTextMapPropagator()
	if len(p.Fields()) == 0 {
		otel.SetTextMapPropagator(GetDefaultTextMapPropagator())
	}
}

// GetDefaultTextMapPropagator returns the default propagator for context propagation between peers.
func GetDefaultTextMapPropagator() propagation.TextMapPropagator {
	return defaultTextMapPropagator
}

// GetTraceID retrieves and returns TraceId from context.
// It returns an empty string is tracing feature is not activated.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID := trace.SpanContextFromContext(ctx).TraceID()
	if traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

// GetSpanID retrieves and returns SpanId from context.
// It returns an empty string is tracing feature is not activated.
func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanID := trace.SpanContextFromContext(ctx).SpanID()
	if spanID.IsValid() {
		return spanID.String()
	}
	return ""
}

// SetBaggageValue is a convenient function for adding one key-value pair to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageValue(ctx context.Context, key string, value interface{}) context.Context {
	return NewBaggage(ctx).SetValue(key, value)
}

// SetBaggageMap is a convenient function for adding map key-value pairs to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageMap(ctx context.Context, data map[string]interface{}) context.Context {
	return NewBaggage(ctx).SetMap(data)
}

// GetBaggageMap retrieves and returns the baggage values as map.
func GetBaggageMap(ctx context.Context) *gmap.StrAnyMap {
	return NewBaggage(ctx).GetMap()
}

// GetBaggageVar retrieves value and returns a *gvar.Var for specified key from baggage.
func GetBaggageVar(ctx context.Context, key string) *gvar.Var {
	return NewBaggage(ctx).GetVar(key)
}

// WithTraceID injects custom trace id into context to propagate.
func WithTraceID(ctx context.Context, traceID string) (context.Context, error) {
	generatedTraceID, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		return ctx, gerror.WrapCodef(
			gcode.CodeInvalidParameter,
			err,
			`invalid custom traceID "%s", a traceID string should be composed with [0-f] and fixed length 32`,
			traceID,
		)
	}

	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		var span trace.Span
		ctx, span = NewSpan(ctx, "gtrace.WithTraceID")
		defer span.End()
		sc = trace.SpanContextFromContext(ctx)
	}
	ctx = trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    generatedTraceID,
		SpanID:     sc.SpanID(),
		TraceFlags: sc.TraceFlags(),
		TraceState: sc.TraceState(),
		Remote:     sc.IsRemote(),
	}))
	return ctx, nil
}

// Tracer warps trace.Tracer for compatibility and extension.
type Tracer struct {
	trace.Tracer
}

// NewTracer Tracer is a short function for retrieving Tracer.
func NewTracer(name ...string) *Tracer {
	tracerName := ""
	if len(name) > 0 {
		tracerName = name[0]
	}
	return &Tracer{
		Tracer: otel.Tracer(tracerName),
	}
}
