package trace

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/chenquan/zap-plus/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrUnknownExporter = errors.New("unknown exporter error")
	tracer             = otel.Tracer("zap-plus")
	agents             = make(map[string]struct{})
	lock               sync.Mutex
	// GetTracerProvider is alis of otel.GetTracerProvider
	GetTracerProvider = otel.GetTracerProvider
)

// StartAgent starts a opentelemetry agent.
func StartAgent(c *config.Trace) {
	lock.Lock()
	defer lock.Unlock()

	_, ok := agents[c.Endpoint]
	if ok {
		return
	}

	// if error happens, let later calls run.
	if err := startAgent(c); err != nil {
		return
	}

	tracer = otel.Tracer(c.Name)

	agents[c.Endpoint] = struct{}{}
}

func startAgent(c *config.Trace) error {
	opts := []sdktrace.TracerProviderOption{
		// Set the sampling rate based on the parent span to 100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(c.Name))),
	}

	if len(c.Endpoint) > 0 {
		exp, err := createExporter(c.Endpoint)
		if err != nil {
			log.Println("opentelemetry exporter err", err)
			return err
		}

		// Always be sure to batch in production.
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Println("opentelemetry error", err)
	}))

	return nil
}

func createExporter(endpoint string) (sdktrace.SpanExporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
}

// Start creates a span and a context.Context containing the newly-created span.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}
