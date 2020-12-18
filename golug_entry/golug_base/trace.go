package golug_base

import (
	"expvar"
	"github.com/pubgo/dix/dix_trace"
)

func (t *baseEntry) trace() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish(t.Options().Name, expvar.Func(func() interface{} { return t.cfg }))
	})
}
