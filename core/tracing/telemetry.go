package tracing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/pubgo/lava/v2/core/lifecycle"
)

var logs = log.GetLogger("tracing")

type Provider struct {
	TracerProvider oteltrace.TracerProvider
	Tracer         oteltrace.Tracer
	MeterProvider  otelmetric.MeterProvider
	Meter          otelmetric.Meter
}

// NewProvider initializes OpenTelemetry globals and registers lifecycle shutdown hooks.
// When tracing is disabled (no exporter endpoint), a noop provider is returned.
func NewProvider(cfg *Config, lc lifecycle.Lifecycle) Provider {
	defer recovery.Exit()

	config := normalizeConfig(cfg)
	if !config.enabled() {
		return noopProvider()
	}

	otel.SetErrorHandler(errorHandler{})

	tracerProvider := newTracerProvider(&config).Unwrap()
	meterProvider := newMeterProvider(&config).Unwrap()
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	if lc != nil {
		lc.AfterStop(func(ctx context.Context) error {
			return errors.Join(
				tracerProvider.Shutdown(ctx),
				meterProvider.Shutdown(ctx),
			)
		})
	}

	return Provider{
		TracerProvider: tracerProvider,
		Tracer: otel.Tracer(
			version.Project(),
			oteltrace.WithInstrumentationVersion(version.Version()),
		),
		MeterProvider: meterProvider,
		Meter: otel.Meter(
			version.Project(),
			otelmetric.WithInstrumentationVersion(version.Version()),
		),
	}
}

func noopProvider() Provider {
	tp := tracenoop.NewTracerProvider()
	mp := metricnoop.NewMeterProvider()
	return Provider{
		TracerProvider: tp,
		Tracer:         tp.Tracer(version.Project()),
		MeterProvider:  mp,
		Meter:          mp.Meter(version.Project()),
	}
}

func mergeResource(config *Config) (r result.Result[*resource.Resource]) {
	defer result.Recovery(&r)
	res := result.Wrap(resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOSType(),
		resource.WithProcessCommandArgs(),
	)).Unwrap()
	res = result.Wrap(resource.Merge(resource.Default(), res)).Unwrap()

	hostname, _ := os.Hostname()
	serviceResource := resource.NewSchemaless(
		semconv.ServiceNameKey.String(config.Service.Name),
		semconv.HostNameKey.String(hostname),
		semconv.ServiceNamespaceKey.String(config.Service.Namespace),
		semconv.ServiceVersionKey.String(config.Service.Version),
		semconv.ProcessPIDKey.Int(os.Getpid()),
		semconv.ProcessCommandKey.String(os.Args[0]),
	)
	res = result.Wrap(resource.Merge(serviceResource, res)).Log().Unwrap()

	return r.WithValue(res)
}

func newTracerProvider(config *Config) (r result.Result[*sdktrace.TracerProvider]) {
	defer result.Recovery(&r)
	res := mergeResource(config).Unwrap()

	traceExporter := result.Wrap(newTraceExporter(config)).Unwrap()
	sampler := sdktrace.ParentBased(sdktrace.AlwaysSample())
	if config.SampleRatio < 1 && config.SampleRatio >= 0 {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SampleRatio))
		log.Info().Msgf("set sample ratio %v", config.SampleRatio)
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

	return r.WithValue(traceProvider)
}

func newTraceExporter(config *Config) (sdktrace.SpanExporter, error) {
	endpoint := config.TraceExporter.ExporterEndpoint
	if endpoint == DefaultStdout {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	if endpoint == "" {
		return nil, fmt.Errorf("trace exporter endpoint is empty")
	}

	traceSecureOption := otlptracegrpc.WithTLSCredentials(config.TraceExporter.Creds)
	if config.TraceExporter.Insecure {
		traceSecureOption = otlptracegrpc.WithInsecure()
	}

	return otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(endpoint),
			traceSecureOption,
			otlptracegrpc.WithHeaders(config.Headers),
			otlptracegrpc.WithCompressor(gzip.Name),
		),
	)
}

func newMeterProvider(config *Config) (r result.Result[*sdkmetric.MeterProvider]) {
	defer result.Recovery(&r)

	exporter := result.Wrap(otelprom.New()).Unwrap()
	res := mergeResource(config).Unwrap()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)
	return r.WithValue(provider)
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

func GetTraceId(ctx context.Context) string {
	return oteltrace.SpanContextFromContext(ctx).TraceID().String()
}
