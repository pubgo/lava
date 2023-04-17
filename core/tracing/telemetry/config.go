package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	traceExporter      *Exporter
	metricExporter     *Exporter
	metricReportPeriod string
	serviceInfo        *ServiceInfo
	attributes         map[string]string
	headers            map[string]string
	idGenerator        sdktrace.IDGenerator
	resource           *resource.Resource
	otelErrorHandler   otel.ErrorHandler
	traceBatchOptions  []sdktrace.BatchSpanProcessorOption
	sampleRatio        float64
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
