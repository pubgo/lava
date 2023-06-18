package metric

import (
	"context"
	"github.com/rs/xid"
	"github.com/uber-go/tally/v4"
)

var metricKey = xid.New().String()

func CreateCtx(parent context.Context, scope tally.Scope) context.Context {
	return context.WithValue(parent, metricKey, scope)
}

func Ctx(ctx context.Context) tally.Scope {
	l, ok := ctx.Value(metricKey).(tally.Scope)
	if ok {
		return l
	}

	return tally.NoopScope
}
