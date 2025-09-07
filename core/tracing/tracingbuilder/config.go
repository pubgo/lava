package tracingbuilder

import (
	"crypto/tls"
	"time"

	"github.com/pubgo/funk/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/durationpb"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	//"google.golang.org/grpc/credentials"
)

const (
	DefaultStdout = "stdout"
)

type TraceConfigLoader struct {
	TraceCfg *Config `yaml:"tracing"`
}

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

// OTLP contains specific configuration used by the OpenTelemetry Metrics exporter.
type OTLP struct {
	GRPC *OtelGRPC `description:"gRPC configuration for the OpenTelemetry collector." json:"grpc,omitempty" toml:"grpc,omitempty" yaml:"grpc,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	HTTP *OtelHTTP `description:"HTTP configuration for the OpenTelemetry collector." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	AddEntryPointsLabels bool                 `description:"Enable metrics on entry points." json:"addEntryPointsLabels,omitempty" toml:"addEntryPointsLabels,omitempty" yaml:"addEntryPointsLabels,omitempty" export:"true"`
	AddRoutersLabels     bool                 `description:"Enable metrics on routers." json:"addRoutersLabels,omitempty" toml:"addRoutersLabels,omitempty" yaml:"addRoutersLabels,omitempty" export:"true"`
	AddServicesLabels    bool                 `description:"Enable metrics on services." json:"addServicesLabels,omitempty" toml:"addServicesLabels,omitempty" yaml:"addServicesLabels,omitempty" export:"true"`
	ExplicitBoundaries   []float64            `description:"Boundaries for latency metrics." json:"explicitBoundaries,omitempty" toml:"explicitBoundaries,omitempty" yaml:"explicitBoundaries,omitempty" export:"true"`
	PushInterval         *durationpb.Duration `description:"Period between calls to collect a checkpoint." json:"pushInterval,omitempty" toml:"pushInterval,omitempty" yaml:"pushInterval,omitempty" export:"true"`
	ServiceName          string               `description:"OTEL service name to use." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (o *OTLP) SetDefaults() {
	o.HTTP = &OtelHTTP{}
	o.HTTP.SetDefaults()

	o.AddEntryPointsLabels = true
	o.AddServicesLabels = true
	o.ExplicitBoundaries = []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	o.PushInterval = durationpb.New(10 * time.Second)
	o.ServiceName = version.Project()
}

// OtelGRPC provides configuration settings for the gRPC open-telemetry.
type OtelGRPC struct {
	Endpoint string            `description:"Sets the gRPC endpoint (host:port) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Insecure bool              `description:"Disables client transport security for the exporter." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (c *OtelGRPC) SetDefaults() {
	c.Endpoint = "localhost:4317"
}

// OtelHTTP provides configuration settings for the HTTP open-telemetry.
type OtelHTTP struct {
	Endpoint string            `description:"Sets the HTTP endpoint (scheme://host:port/path) of the collector." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Headers  map[string]string `description:"Headers sent with payload." json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
}

// SetDefaults sets the default values.
func (c *OtelHTTP) SetDefaults() {
	c.Endpoint = "https://localhost:4318"
}
