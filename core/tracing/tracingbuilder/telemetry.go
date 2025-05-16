package tracingbuilder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/core/lifecycle"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/encoding/gzip"
)

type Provider struct {
	TracerProvider oteltrace.TracerProvider
	Tracer         oteltrace.Tracer
	MeterProvider  otelmetric.MeterProvider
	Meter          otelmetric.Meter
}

type Params struct {
	Cfg *Config
	LC  lifecycle.Lifecycle
}

func New(params Params) Provider {
	config := &Config{
		traceExporter:      &Exporter{},
		metricExporter:     &Exporter{},
		metricReportPeriod: "",
		serviceInfo:        &ServiceInfo{},
		attributes:         map[string]string{},
		headers:            map[string]string{},
		idGenerator:        nil,
		traceBatchOptions:  []sdktrace.BatchSpanProcessorOption{},
		sampleRatio:        1,
	}

	tracerProvider := NewTracerProvider(config)
	meterProvider := NewMeterProvider(config)
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	params.LC.AfterStop(func(ctx context.Context) error {
		return errors.Join(
			tracerProvider.Shutdown(ctx),
			meterProvider.Shutdown(ctx),
		)
	})

	return Provider{
		TracerProvider: tracerProvider,
		Tracer: otel.Tracer(
			version.Project(),
			oteltrace.WithInstrumentationVersion(version.Version()),
			oteltrace.WithInstrumentationAttributes(),
		),
		MeterProvider: meterProvider,
		Meter: otel.Meter(
			version.Project(),
			otelmetric.WithInstrumentationVersion(version.Version()),
			otelmetric.WithInstrumentationAttributes(),
		),
	}
}

// merge config resource with default resource
func mergeResource(config *Config) *resource.Resource {
	res := assert.Must1(resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOSType(),
		resource.WithProcessCommandArgs(),
	))
	res = assert.Must1(resource.Merge(resource.Default(), res))

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
	res = assert.Must1(resource.Merge(resource.Default(), defaultResource))

	return res
}

func NewTracerProvider(config *Config) *sdktrace.TracerProvider {
	res := mergeResource(config)

	traceExporter := assert.Must1(newGrpcTracerExporter(config))
	sampler := sdktrace.ParentBased(sdktrace.AlwaysSample())
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

func newGrpcTracerExporter(config *Config) (sdktrace.SpanExporter, error) {
	if config.traceExporter.ExporterEndpoint == DefaultStdout {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	//opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	traceSecureOption := otlptracegrpc.WithTLSCredentials(config.traceExporter.Creds)
	if config.traceExporter.Insecure {
		traceSecureOption = otlptracegrpc.WithInsecure()
	}

	return otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(config.traceExporter.ExporterEndpoint),
			traceSecureOption,
			otlptracegrpc.WithHeaders(config.headers),
			otlptracegrpc.WithCompressor(gzip.Name),
		),
	)
}

func newGrpcMetricExporter(config *Config) (metric.Exporter, error) {
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

func NewMeterProvider(config *Config) *sdkmetric.MeterProvider {
	reader := metric.NewPeriodicReader(assert.Must1(newGrpcMetricExporter(config)))
	readerOpt := sdkmetric.WithReader(reader)

	exporter := assert.Must1(otelprom.New())
	readerOpt = sdkmetric.WithReader(exporter)

	res := mergeResource(config)
	provider := sdkmetric.NewMeterProvider(
		readerOpt,
		sdkmetric.WithResource(res),
	)
	return provider
}
