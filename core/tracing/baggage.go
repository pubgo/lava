package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
)

const (
	Baggage                 = "baggage"
	AttributeTraceID        = attribute.Key("trace.id")
	AttributeSpanID         = attribute.Key("span.id")
	AttributeLogID          = attribute.Key("log_id")
	AttributeRequest        = attribute.Key("request")
	AttributeResponse       = attribute.Key("response")
	AttributeGinError       = attribute.Key("gin.errors")
	AttributeRedisError     = attribute.Key("redis.cmd.error")
	AttributeRedisCmdName   = attribute.Key("redis.cmd.name")
	AttributeRedisCmdString = attribute.Key("redis.cmd.string")
	AttributeRedisCmdArgs   = attribute.Key("redis.cmd.args")
)

func ExtractHTTPBaggage(ctx context.Context, header http.Header) context.Context {
	b, err := baggage.Parse(header.Get(Baggage))
	if err != nil {
		return ctx
	}

	ctx = baggage.ContextWithBaggage(ctx, b)
	if header == nil {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

func InjectHTTPBaggage(ctx context.Context, header http.Header) {
	if header == nil {
		return
	}

	header.Set(Baggage, baggage.FromContext(ctx).String())
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}
