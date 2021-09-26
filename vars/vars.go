package vars

import (
	"expvar"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/runenv"
	"runtime/debug"

	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

type value func() interface{}

func (f value) Value() interface{} { return f() }
func (f value) String() (val string) {
	xerror.TryCatch(func() {
		dt := f()
		v := xerror.PanicBytes(jsonx.Marshal(dt))
		val = byteutil.ToStr(v)
	}, func(err error) {
		val = jsonx.Json(typex.M{
			"err_msg": err,
			"err":     err.Error(),
			"stack":   string(debug.Stack()),
		})

		if runenv.IsDev() {
			xerror.Parse(err).Debug()
		}
	})

	return
}

func Watch(name string, data func() interface{}) {
	expvar.Publish(name, value(data))
}

func Get(name string) expvar.Var {
	return expvar.Get(name)
}

func Each(fn func(key string, val func() string)) {
	expvar.Do(func(kv expvar.KeyValue) { fn(kv.Key, kv.Value.String) })
}
