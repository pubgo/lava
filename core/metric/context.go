package metric

import (
	"context"

	"github.com/uber-go/tally"
)

type metricKey struct{}

func CreateCtx(parent context.Context, scope tally.Scope) context.Context {
	return context.WithValue(parent, metricKey{}, scope)
}

func GetMetric(ctx context.Context) tally.Scope {
	var l, ok = ctx.Value(metricKey{}).(tally.Scope)
	if ok {
		return l
	}
	return tally.NoopScope
}
