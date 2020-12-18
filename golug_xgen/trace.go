package golug_xgen

import (
	"expvar"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish("xgen", expvar.Func(func() interface{} {
			dt := make(map[string][]GrpcRestHandler)
			data.Range(func(key, value interface{}) bool {
				var _e = runtime.FuncForPC(key.(reflect.Value).Pointer())
				var file, line = _e.FileLine(key.(reflect.Value).Pointer())

				var buf = &strings.Builder{}
				defer buf.Reset()

				buf.WriteString(file)
				buf.WriteString(":")
				buf.WriteString(strconv.Itoa(line))
				buf.WriteString(" ")

				ma := strings.Split(_e.Name(), ".")
				buf.WriteString(ma[len(ma)-1])
				dt[buf.String()] = value.([]GrpcRestHandler)
				return true
			})
			return dt
		}))
	})
}
