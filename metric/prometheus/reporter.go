package prometheus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pubgo/golug/metric"
	"go.uber.org/atomic"
)

var (
	// quantileThresholds maps quantiles/percentiles to error thresholds (required by the Prometheus client).
	// Must be from our pre-defined set [0.0, 0.5, 0.75, 0.90, 0.95, 0.98, 0.99, 1]:
	quantileThresholds = map[float64]float64{0.0: 0, 0.5: 0.05, 0.75: 0.04, 0.90: 0.03, 0.95: 0.02, 0.98: 0.001, 1: 0}
)

var _ metric.Reporter = (*Reporter)(nil)

// Reporter is an implementation of metrics.Reporter:
type Reporter struct {
	cfg                Cfg
	metrics            metricFamily
	server             *http.Server
	prometheusRegistry *prometheus.Registry
	isStarted          atomic.Bool
}

func (r *Reporter) String() string { return Name }

// Start ...
func (r *Reporter) Start() (err error) {
	// Handle the metrics endpoint with prometheus:
	fmt.Printf("Metrics/Prometheus [http] Listening on %s%s\n", r.cfg.Address, r.cfg.Path)
	go func() {
		r.isStarted.Store(true)
		r.server.Handler = metric.DefaultServeMux
		if err := r.server.ListenAndServe(); err == http.ErrServerClosed {
			log.Println(err)
		} else {
			log.Fatalln(err)
		}
		r.isStarted.Store(false)
	}()
	time.Sleep(time.Second)
	return nil
}

// Stop ...
func (r *Reporter) Stop() error {
	if !r.isStarted.Load() {
		return MetricStartError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.server.Shutdown(ctx)
}

// NewReporter ...
func NewReporter(opts ...metric.Option) (metric.Reporter, error) {
	return newReporter(opts...)
}

// newReporter returns a configured prometheus reporter:
func newReporter(opts ...metric.Option) (reporter *Reporter, err error) {
	defer errHandler(&err)

	options := metric.NewOptions(opts...)

	options.Name = metric.StripUnsupportedCharacters(strings.ToLower(strings.TrimSpace(options.Name)))
	if options.Name != "" && !strings.HasSuffix(options.Name, "_") {
		options.Name += "_"
	}

	options.Path = strings.ToLower(strings.TrimSpace(options.Path))
	if options.Path == "" {
		options.Path = "/metrics"
	}

	// Make a prometheus registry (this keeps track of any metrics we generate):
	prometheusRegistry := prometheus.NewRegistry()
	if err = prometheusRegistry.Register(prometheus.NewGoCollector()); err != nil {
		return nil, err
	}
	if err = prometheusRegistry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{Namespace: "go"})); err != nil {
		return nil, err
	}

	// Make a new Reporter:
	reporter = &Reporter{
		cfg:                options,
		prometheusRegistry: prometheusRegistry,
		server:             &http.Server{Addr: options.Address},
	}

	// Add metrics families for each type:
	reporter.metrics = reporter.newMetricFamily()

	metric.DefaultServeMux.Handle(options.Path, promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	return reporter, nil
}

// convertTags turns Tags into prometheus labels:
func (r *Reporter) convertTags(tags metric.Tags) prometheus.Labels {
	labels := prometheus.Labels{}
	for key, value := range tags {
		labels[key] = metric.StripUnsupportedCharacters(value)
	}
	return labels
}

// listTagKeys returns a list of tag keys (we need to provide this to the Prometheus client):
func listTagKeys(tags metric.Tags) (labelKeys []string) {
	for key := range tags {
		labelKeys = append(labelKeys, key)
	}
	return
}
