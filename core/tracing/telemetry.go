package tracing

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpTraceGrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	otelmetric "go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/encoding/gzip"

	"github.com/pubgo/lava/core/lifecycle"
)

const (
	DefaultStdout = "stdout"
)

type Provider struct {
	TracerProvider oteltrace.TracerProvider
	Tracer         oteltrace.Tracer
	MeterProvider  otelmetric.MeterProvider
	Meter          otelmetric.Meter
}

func New(cfg *Config, lc lifecycle.Lifecycle) Provider {
	config := &Config{
		traceExporter:      &Exporter{},
		metricExporter:     &Exporter{},
		metricReportPeriod: "",
		serviceInfo:        &ServiceInfo{},
		attributes:         map[string]string{},
		headers:            map[string]string{},
		idGenerator:        nil,
		otelErrorHandler:   errorHandler{},
		traceBatchOptions:  []sdktrace.BatchSpanProcessorOption{},
		sampleRatio:        1,
	}

	otel.SetErrorHandler(errorHandler{})

	tracerProvider := NewTracer(config)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	meterProvider := NewPrometheusMeterProvider(config)

	lc.AfterStop(func() {
		assert.Must(tracerProvider.Shutdown(context.Background()))
		assert.Must(meterProvider.Shutdown(context.Background()))
	})

	return Provider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}
}

type errorHandler struct {
}

// Handle default error handler when span send failed
func (errorHandler) Handle(err error) {
	log.Err(err).Msg("tracer exporter error")
}

// merge config resource with default resource
func mergeResource(config *Config) *resource.Resource {
	hostname, _ := os.Hostname()
	defaultResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.serviceInfo.Name),
		semconv.HostNameKey.String(hostname),
		semconv.ServiceNamespaceKey.String(config.serviceInfo.Namespace),
		semconv.ServiceVersionKey.String(config.serviceInfo.Version),
		semconv.ProcessPIDKey.Int(os.Getpid()),
		semconv.ProcessCommandKey.String(os.Args[0]),
	)

	return assert.Must1(resource.Merge(resource.Default(), defaultResource))
}

func NewTracer(config *Config) *sdktrace.TracerProvider {
	res := mergeResource(config)

	traceExporter := assert.Must1(initTracerExporter(config))
	sampler := sdktrace.AlwaysSample()
	if config.sampleRatio < 1 && config.sampleRatio >= 0 {
		sampler = sdktrace.TraceIDRatioBased(config.sampleRatio)
		log.Info().Msgf("set sample ratio %v", config.sampleRatio)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithMaxQueueSize(queueSize()),
			sdktrace.WithMaxExportBatchSize(queueSize()),
			sdktrace.WithBatchTimeout(10*time.Second),
			sdktrace.WithExportTimeout(10*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	return traceProvider
}

func initTracerExporter(config *Config) (sdktrace.SpanExporter, error) {
	if config.traceExporter.ExporterEndpoint == DefaultStdout {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if config.traceExporter.ExporterEndpoint != "" {
		traceSecureOption := otlpTraceGrpc.WithTLSCredentials(config.traceExporter.Creds)
		if config.traceExporter.Insecure {
			traceSecureOption = otlpTraceGrpc.WithInsecure()
		}

		return otlptrace.New(
			context.Background(),
			otlpTraceGrpc.NewClient(
				otlpTraceGrpc.WithEndpoint(config.traceExporter.ExporterEndpoint),
				traceSecureOption,
				otlpTraceGrpc.WithHeaders(config.headers),
				otlpTraceGrpc.WithCompressor(gzip.Name),
			),
		)
	}

	return nil, fmt.Errorf("tracer exporter endpoint is nil, no exporter is inited")
}

func NewPrometheusMeterProvider(config *Config) *sdkmetric.MeterProvider {
	res := mergeResource(config)
	exporter := assert.Must1(otelprom.New())
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)
	return provider
}

// GetTraceId return trace id in context
func GetTraceId(ctx context.Context) string {
	return oteltrace.SpanContextFromContext(ctx).TraceID().String()
}

func queueSize() int {
	const min = 1000
	const max = 16000

	n := (runtime.GOMAXPROCS(0) / 2) * 1000
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
