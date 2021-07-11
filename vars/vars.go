package vars

import (
	"expvar"

	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"
)

var logs = xlog.GetLogger("vars")

type value func() interface{}

func (f value) Value() interface{} { return f() }
func (f value) String() (val string) {
	var dt interface{}
	try.Logs(logs, func() { dt = f() })

	v, err := jsonx.Marshal(dt)
	if err != nil {
		return err.Error()
	}

	return byteutil.ToStr(v)
}

func Watch(name string, data func() interface{}) {
	expvar.Publish(name, value(func() (val interface{}) {
		defer xerror.Resp(func(err xerror.XErr) {
			logs.Error("unknown error",
				zap.String("pkg", "expvar"),
				zap.String("name", name),
				zap.Any("err", err))
		})
		return data()
	}))
}

func Get(name string) expvar.Var {
	return expvar.Get(name)
}

func Each(fn func(key string, val func() string)) {
	expvar.Do(func(kv expvar.KeyValue) { fn(kv.Key, kv.Value.String) })
}
