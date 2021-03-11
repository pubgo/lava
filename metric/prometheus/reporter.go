package prometheus

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/metric"
	"github.com/pubgo/golug/mux"
	"github.com/pubgo/xerror"
)

var (
	// quantileThresholds maps quantiles/percentiles to error thresholds (required by the Prometheus client).
	// Must be from our pre-defined set [0.0, 0.5, 0.75, 0.90, 0.95, 0.98, 0.99, 1]:
	quantileThresholds = map[float64]float64{0.0: 0, 0.5: 0.05, 0.75: 0.04, 0.90: 0.03, 0.95: 0.02, 0.98: 0.001, 1: 0}
)

func init() {
	xerror.Exit(metric.Register(Name, New))
}

func New(cfg map[string]interface{}) (r metric.Reporter, err error) {
	defer xerror.RespErr(&err)

	var cfg1 = GetDefaultCfg()
	xerror.Exit(gutils.Decode(cfg, &cfg1))

	return newReporter(cfg1)
}

var _ metric.Reporter = (*Reporter)(nil)

// Reporter is an implementation of metrics.Reporter:
type Reporter struct {
	cfg                Cfg
	metrics            metricFamily
	prometheusRegistry *prometheus.Registry
}

func (r *Reporter) Name() string { return Name }

// newReporter returns a configured prometheus reporter:
func newReporter(cfg Cfg) (reporter *Reporter, err error) {
	defer xerror.RespErr(&err)

	var name = cfg.Prefix
	name = StripUnsupportedCharacters(strings.ToLower(strings.TrimSpace(name)))
	if name != "" && !strings.HasSuffix(name, "_") {
		name += "_"
	}
	cfg.Prefix = name

	cfg.Path = strings.ToLower(strings.TrimSpace(cfg.Path))
	if cfg.Path == "" {
		cfg.Path = "/metrics"
	}

	// Make a prometheus registry (this keeps track of any metrics we generate):
	prometheusRegistry := prometheus.NewRegistry()
	xerror.Panic(prometheusRegistry.Register(prometheus.NewGoCollector()))
	xerror.Panic(prometheusRegistry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{Namespace: "go"})))

	// Make a new Reporter:
	reporter = &Reporter{cfg: cfg, prometheusRegistry: prometheusRegistry}

	// Add metrics families for each type:
	reporter.metrics = reporter.newMetricFamily()
	mux.Default().Handle(cfg.Path,
		promhttp.HandlerFor(prometheusRegistry,
			promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))

	return reporter, nil
}

// convertTags turns Tags into prometheus labels:
func (r *Reporter) convertTags(tags metric.Tags) prometheus.Labels {
	labels := prometheus.Labels{}
	for key, value := range tags {
		labels[key] = StripUnsupportedCharacters(value)
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
