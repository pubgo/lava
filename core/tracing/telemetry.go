package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
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

	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
)

var logs = log.GetLogger("tracing")

const (
	DefaultStdout = "stdout"

	grpcHealthyMethod   = "/healthy.HealthService/Health"
	healthHost          = "127.0.0.1"
	instrumentationName = "github.com/gowins/dionysus/opentelemetry"
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

	//name := instrumentationName + "/" + config.serviceInfo.Namespace + "/" + config.serviceInfo.Name
	//	defaultTracer = otel.GetTracerProvider().Tracer(name, oteltrace.WithInstrumentationVersion("v1.1.0"))

	return Provider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}
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
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.sampleRatio))
		log.Info().Msgf("set sample ratio %v", config.sampleRatio)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithMaxQueueSize(queueSize()),
			sdktrace.WithMaxExportBatchSize(queueSize()),
			sdktrace.WithBatchTimeout(10*time.Second),
			sdktrace.WithExportTimeout(10*time.Second),
		),
		sdktrace.WithRawSpanLimits(sdktrace.SpanLimits{
			AttributeCountLimit:         1024,
			EventCountLimit:             1024,
			LinkCountLimit:              1024,
			AttributePerEventCountLimit: 1024,
			AttributePerLinkCountLimit:  1024,
			AttributeValueLengthLimit:   1024,
		}),
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

func TraceID(span oteltrace.Span) string {
	traceID := span.SpanContext().TraceID()
	if traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

func SpanID(span oteltrace.Span) string {
	spanID := span.SpanContext().SpanID()
	if spanID.IsValid() {
		return spanID.String()
	}
	return ""
}

func TraceIdFromCtx(ctx context.Context) string {
	return oteltrace.SpanContextFromContext(ctx).TraceID().String()
}

func Tracer() oteltrace.Tracer {
	return otel.Tracer(version.Project())
}

func CheckHasTraceID(ctx context.Context) bool {
	return oteltrace.SpanFromContext(ctx).SpanContext().HasTraceID()
}

// GetTraceId return trace id in context
func GetTraceId(ctx context.Context) string {
	return oteltrace.SpanContextFromContext(ctx).TraceID().String()
}

func initMetricExporter(config *Config) (metric.Exporter, error) {
	if config.metricExporter.ExporterEndpoint == DefaultStdout {
		encoder := json.NewEncoder(os.Stdout)
		return stdoutmetric.New(stdoutmetric.WithEncoder(encoder))
	}

	if config.metricExporter.ExporterEndpoint != "" {
		metricSecureOption := otlpmetricgrpc.WithTLSCredentials(config.metricExporter.Creds)
		if config.metricExporter.Insecure {
			metricSecureOption = otlpmetricgrpc.WithInsecure()
		}

		return otlpmetricgrpc.New(
			context.Background(),
			otlpmetricgrpc.WithEndpoint(config.metricExporter.ExporterEndpoint),
			metricSecureOption,
			otlpmetricgrpc.WithHeaders(config.headers),
			otlpmetricgrpc.WithCompressor(gzip.Name))
	}
	return nil, fmt.Errorf("metric exporter endpoint is nil, no exporter is inited")
}
