package metrics

import (
	"context"

	"github.com/rs/xid"
	"github.com/uber-go/tally/v4"
)

type ctxKet string

var metricKey = ctxKet(xid.New().String())

func InjectToCtx(parent context.Context, scope tally.Scope) context.Context {
	return context.WithValue(parent, metricKey, scope)
}

func GetFromCtx(ctx context.Context) tally.Scope {
	l, ok := ctx.Value(metricKey).(tally.Scope)
	if ok {
		return l
	}

	return tally.NoopScope
}
