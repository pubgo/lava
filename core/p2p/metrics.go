package p2p

import (
	"time"

	"github.com/pubgo/lava/v2/core/metrics"
)

const metricsPrefix = "p2p"

// MetricsRecorder 上报 P2P 建连指标（scope 为 nil 时为 noop）。
type MetricsRecorder struct {
	scope metrics.Metric
}

// NewMetricsRecorder 创建指标记录器；scope 为 nil 时返回 nil（noop）。
func NewMetricsRecorder(scope metrics.Metric) *MetricsRecorder {
	if scope == nil {
		return nil
	}
	return &MetricsRecorder{scope: scope}
}

func pathType(pair CandidatePairInfo) string {
	if pair.LocalType == "relay" || pair.RemoteType == "relay" {
		return "relay"
	}
	if pair.LocalType == "srflx" || pair.RemoteType == "srflx" {
		return "srflx"
	}
	return "host"
}

// ObserveConnect 记录一次成功建连。
func (m *MetricsRecorder) ObserveConnect(pair CandidatePairInfo, d time.Duration) {
	if m == nil || m.scope == nil {
		return
	}
	pt := pathType(pair)
	tagged := m.scope.Tagged(metrics.Tags{
		"path":        pt,
		"local_type":  pair.LocalType,
		"remote_type": pair.RemoteType,
	})
	tagged.Counter(metricsPrefix + ".connect_total").Inc(1)
	tagged.Timer(metricsPrefix + ".connect_duration").Record(d)
	if pt == "relay" {
		tagged.Counter(metricsPrefix + ".relay_connect_total").Inc(1)
	}
}

// ObserveDialFailure 记录一次拨号失败。
func (m *MetricsRecorder) ObserveDialFailure() {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Counter(metricsPrefix + ".dial_failure_total").Inc(1)
}

// SetActiveConnections 更新当前活跃连接数与中继连接数。
func (m *MetricsRecorder) SetActiveConnections(active, relay int) {
	if m == nil || m.scope == nil {
		return
	}
	m.scope.Gauge(metricsPrefix + ".connections_active").Update(float64(active))
	m.scope.Gauge(metricsPrefix + ".relay_connections_active").Update(float64(relay))
}
