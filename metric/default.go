package metric

import (
	"github.com/pubgo/xerror"
)

var defaultReporter Reporter
// GetDefault 获取全局的Reporter
func GetDefault() Reporter {
	xerror.Assert(defaultReporter == nil, "please set default reporter")
	return defaultReporter
}

// Count 上报递增数据
func Count(name string, value float64, tags Tags) error {
	return GetDefault().Count(name, value, tags)
}

// Gauge 实时的上报当前指标
func Gauge(name string, value float64, tags Tags) error {
	return GetDefault().Gauge(name, value, tags)
}

// Histogram 存储区间数据, 在服务端端聚合数据
func Histogram(name string, value float64, tags Tags, opts *HistogramOpts) error {
	return GetDefault().Histogram(name, value, tags, opts)
}

// Summarier 在 client 端聚合数据, 直接存储了分位数
func Summary(name string, value float64, tags Tags) error {
	return GetDefault().Summary(name, value, tags)
}
