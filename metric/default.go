package metric

import (
	"github.com/pubgo/xerror"
)

var defaultReporter Reporter = &noopReporter{}

// setDefault 设置全局的Reporter
func setDefault(reporter Reporter) {
	xerror.Assert(reporter == nil, "[reporter] should not be nil")
	defaultReporter = reporter
}

// getDefault 获取全局的Reporter
func getDefault() Reporter {
	xerror.Assert(defaultReporter == nil, "please set default reporter")
	return defaultReporter
}

//CreateGauge init a new gauge type
func CreateGauge(name string, labels []string, opts GaugeOpts) error {
	return getDefault().CreateGauge(name, labels, opts)
}

//CreateCounter init a new counter type
func CreateCounter(name string, labels []string, opts CounterOpts) error {
	return getDefault().CreateCounter(name, labels, opts)
}

//CreateSummary init a new summary type
func CreateSummary(name string, labels []string, opts SummaryOpts) error {
	return getDefault().CreateSummary(name, labels, opts)
}

//CreateHistogram init a new histogram type
func CreateHistogram(name string, labels []string, opts HistogramOpts) error {
	return getDefault().CreateHistogram(name, labels, opts)
}

// Count 上报递增数据
func Count(name string, value float64, tags Tags) error {
	return getDefault().Count(name, value, tags)
}

// Gauge 实时的上报当前指标
func Gauge(name string, value float64, tags Tags) error {
	return getDefault().Gauge(name, value, tags)
}

// Histogram 存储区间数据, 在服务端端聚合数据
func Histogram(name string, value float64, tags Tags) error {
	return getDefault().Histogram(name, value, tags)
}

// Summary 在client端聚合数据, 直接存储了分位数
func Summary(name string, value float64, tags Tags) error {
	return getDefault().Summary(name, value, tags)
}
