package prometheus

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/pubgo/golug/metric"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusReporter(t *testing.T) {
	// Make a Reporter:
	reporter, err := newReporter(
		metric.Address(":9999"),
		metric.Path("/prometheus"),
		metric.DefaultTags(map[string]string{"service": "prometheus-test"}),
	)

	defer func() {
		assert.NoError(t, reporter.Stop())
	}()
	assert.NoError(t, reporter.Start())

	assert.NoError(t, err)
	assert.NotNil(t, reporter)
	assert.Equal(t, "prometheus-test", reporter.cfg.DefaultTags["service"])
	assert.Equal(t, ":9999", reporter.cfg.Address)
	assert.Equal(t, "/prometheus", reporter.cfg.Path)

	// Check that our implementation is valid:
	assert.Implements(t, new(metric.Reporter), reporter)

	// Test tag conversion:
	tags := metric.Tags{
		"tag1": "false",
		"tag2": "true",
	}
	convertedTags := reporter.convertTags(tags)
	assert.Equal(t, "false", convertedTags["tag1"])
	assert.Equal(t, "true", convertedTags["tag2"])

	// Test tag enumeration:
	listedTags := listTagKeys(tags)
	assert.Contains(t, listedTags, "tag1")
	assert.Contains(t, listedTags, "tag2")

	// Test string cleaning:
	preparedMetricName := metric.StripUnsupportedCharacters("some.kind,of tag")
	assert.Equal(t, "some_kind_oftag", preparedMetricName)

	// Test MetricFamilies:
	metricFamily := reporter.newMetricFamily()

	// Counters:
	assert.NotNil(t, metricFamily.getCounter("testCounter", metric.Tags{"test": "", "counter": ""}))
	assert.Len(t, metricFamily.counters, 1)

	// Gauges:
	assert.NotNil(t, metricFamily.getGauge("testGauge", metric.Tags{"test": "", "gauge": ""}))
	assert.Len(t, metricFamily.gauges, 1)

	// Timings:
	assert.NotNil(t, metricFamily.getSummary("testTiming", metric.Tags{"test": "", "timing": ""}))
	assert.Len(t, metricFamily.summaries, 1)

	// Test submitting metrics through the interface methods:
	assert.NoError(t, reporter.Count("test.counter.1", 6, tags))
	assert.NoError(t, reporter.Count("test.counter.2", 19, tags))
	assert.NoError(t, reporter.Count("test.counter.1", 5, tags))
	assert.NoError(t, reporter.Gauge("test.gauge.1", 99, tags))
	assert.NoError(t, reporter.Gauge("test.gauge.2", 55, tags))
	assert.NoError(t, reporter.Gauge("test.gauge.1", 98, tags))
	assert.NoError(t, reporter.Summary("test.timing.1", time.Second.Seconds(), tags))
	assert.NoError(t, reporter.Summary("test.timing.2", time.Minute.Seconds(), tags))
	assert.Len(t, reporter.metrics.counters, 2)
	assert.Len(t, reporter.metrics.gauges, 2)
	assert.Len(t, reporter.metrics.summaries, 2)

	// Test reading back the metrics:
	rsp, err := http.Get("http://localhost:9999/prometheus")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rsp.StatusCode)

	// Read the response body and check for our metrics:
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	assert.NoError(t, err)

	// Check for appropriately aggregated metrics:
	assert.Contains(t, string(bodyBytes), `test_counter_1{service="prometheus-test",tag1="false",tag2="true"} 11`)
	assert.Contains(t, string(bodyBytes), `test_counter_2{service="prometheus-test",tag1="false",tag2="true"} 19`)
	assert.Contains(t, string(bodyBytes), `test_gauge_1{service="prometheus-test",tag1="false",tag2="true"} 98`)
	assert.Contains(t, string(bodyBytes), `test_gauge_2{service="prometheus-test",tag1="false",tag2="true"} 55`)
	assert.Contains(t, string(bodyBytes), `test_timing_1{service="prometheus-test",tag1="false",tag2="true",quantile="0"} 1`)
	assert.Contains(t, string(bodyBytes), `test_timing_2{service="prometheus-test",tag1="false",tag2="true",quantile="0"} 60`)
}
