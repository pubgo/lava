package tracing

import (
	"crypto/tls"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/durationpb"
)

const DefaultStdout = "stdout"

type TraceConfigLoader struct {
	TraceCfg *Config `yaml:"tracing"`
}

type Config struct {
	TraceExporter  Exporter          `yaml:"trace_exporter"`
	MetricExporter Exporter          `yaml:"metric_exporter"`
	SampleRatio    float64           `yaml:"sample_ratio"`
	Service        ServiceInfo       `yaml:"service"`
	Headers        map[string]string `yaml:"headers"`

	metricReportPeriod string
	attributes         map[string]string
	idGenerator        sdktrace.IDGenerator
	otelErrorHandler   otel.ErrorHandler
	traceBatchOptions  []sdktrace.BatchSpanProcessorOption

	resourceAttributes []attribute.KeyValue
	resourceDetectors  []resource.Detector
	tlsConf            *tls.Config

	tracingEnabled    bool
	textMapPropagator propagation.TextMapPropagator
	tracerProvider    *sdktrace.TracerProvider
	traceSampler      sdktrace.Sampler
	prettyPrint       bool
	bspOptions        []sdktrace.BatchSpanProcessorOption

	metricsEnabled bool
	metricOptions  []metric.Option
}

type Exporter struct {
	ExporterEndpoint string                           `yaml:"endpoint"`
	Insecure         bool                             `yaml:"insecure"`
	Creds            credentials.TransportCredentials `yaml:"-"`
}

type ServiceInfo struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Version   string `yaml:"version"`
}

// OTLP contains specific configuration used by the OpenTelemetry Metric exporter.
type OTLP struct {
	GRPC *OtelGRPC `yaml:"grpc,omitempty"`
	HTTP *OtelHTTP `yaml:"http,omitempty"`

	AddEntryPointsLabels bool                 `yaml:"addEntryPointsLabels,omitempty"`
	AddRoutersLabels     bool                 `yaml:"addRoutersLabels,omitempty"`
	AddServicesLabels    bool                 `yaml:"addServicesLabels,omitempty"`
	ExplicitBoundaries   []float64            `yaml:"explicitBoundaries,omitempty"`
	PushInterval         *durationpb.Duration `yaml:"pushInterval,omitempty"`
	ServiceName          string               `yaml:"serviceName,omitempty"`
}

func (o *OTLP) SetDefaults() {
	o.HTTP = &OtelHTTP{}
	o.HTTP.SetDefaults()

	o.AddEntryPointsLabels = true
	o.AddServicesLabels = true
	o.ExplicitBoundaries = []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	o.PushInterval = durationpb.New(10 * time.Second)
	o.ServiceName = version.Project()
}

type OtelGRPC struct {
	Endpoint string            `yaml:"endpoint"`
	Insecure bool              `yaml:"insecure"`
	Headers  map[string]string `yaml:"headers"`
}

func (c *OtelGRPC) SetDefaults() {
	c.Endpoint = "localhost:4317"
}

type OtelHTTP struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"headers"`
}

func (c *OtelHTTP) SetDefaults() {
	c.Endpoint = "https://localhost:4318"
}

func DefaultCfg() Config {
	return Config{
		SampleRatio: 1,
		Service: ServiceInfo{
			Name:      version.Project(),
			Namespace: version.Project(),
			Version:   version.Version(),
		},
		Headers: map[string]string{},
	}
}

func (c *Config) enabled() bool {
	if c == nil {
		return false
	}
	ep := c.TraceExporter.ExporterEndpoint
	return ep == DefaultStdout || ep != ""
}

func normalizeConfig(cfg *Config) Config {
	if cfg == nil {
		return DefaultCfg()
	}
	out := *cfg
	if out.SampleRatio == 0 {
		out.SampleRatio = 1
	}
	if out.Service.Name == "" {
		out.Service.Name = version.Project()
	}
	if out.Service.Version == "" {
		out.Service.Version = version.Version()
	}
	if out.Headers == nil {
		out.Headers = map[string]string{}
	}
	return out
}
