package tracing

import (
	"crypto/tls"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	//"google.golang.org/grpc/credentials"
)

type Config struct {
	traceExporter      *Exporter
	metricExporter     *Exporter
	metricReportPeriod string
	serviceInfo        *ServiceInfo
	attributes         map[string]string
	headers            map[string]string
	idGenerator        sdktrace.IDGenerator
	otelErrorHandler   otel.ErrorHandler
	traceBatchOptions  []sdktrace.BatchSpanProcessorOption
	sampleRatio        float64

	resourceAttributes []attribute.KeyValue
	resourceDetectors  []resource.Detector

	tlsConf *tls.Config

	// Tracing options

	tracingEnabled    bool
	textMapPropagator propagation.TextMapPropagator
	tracerProvider    *sdktrace.TracerProvider
	traceSampler      sdktrace.Sampler
	prettyPrint       bool
	bspOptions        []sdktrace.BatchSpanProcessorOption

	// Metrics options

	metricsEnabled bool
	metricOptions  []metric.Option
}

type Exporter struct {
	ExporterEndpoint string
	Insecure         bool
	Creds            credentials.TransportCredentials
}

type ServiceInfo struct {
	Name      string
	Namespace string
	Version   string
}
