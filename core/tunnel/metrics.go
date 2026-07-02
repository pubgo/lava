package tunnel

import (
	"github.com/pubgo/lava/v2/core/metrics"
)

const metricsPrefix = "tunnel"

// MetricsRecorder 上报 tunnel gateway 指标（scope 为 nil 时为 noop）。
type MetricsRecorder struct {
	scope metrics.Metric
}

// NewMetricsRecorder 创建指标记录器；scope 为 nil 时返回 nil。
func NewMetricsRecorder(scope metrics.Metric) *MetricsRecorder {
	if scope == nil {
		return nil
	}
	return &MetricsRecorder{scope: scope}
}

// ObserveProxyRequest 记录一次代理请求。
func (m *MetricsRecorder) ObserveProxyRequest(service, endpoint string) {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Tagged(metrics.Tags{"service": service, "endpoint": endpoint}).
		Counter(metricsPrefix + ".proxy_total").Inc(1)
}

// ObserveProxyAuthFailure 记录一次代理鉴权失败。
func (m *MetricsRecorder) ObserveProxyAuthFailure() {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Counter(metricsPrefix + ".proxy_auth_failure_total").Inc(1)
}

// ObserveRateLimited 记录一次限流拒绝。
func (m *MetricsRecorder) ObserveRateLimited(service string) {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Tagged(metrics.Tags{"service": service}).
		Counter(metricsPrefix + ".rate_limit_total").Inc(1)
}

// ObserveRegister 记录一次服务注册。
func (m *MetricsRecorder) ObserveRegister(service string) {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Tagged(metrics.Tags{"service": service}).
		Counter(metricsPrefix + ".register_total").Inc(1)
}

// SetRegisteredServices 更新当前注册服务数。
func (m *MetricsRecorder) SetRegisteredServices(n int) {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Gauge(metricsPrefix + ".services_registered").Update(float64(n))
}
