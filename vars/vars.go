package vars

import (
	"expvar"

	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

type value func() interface{}

func (f value) Value() interface{} { return f() }
func (f value) String() string {
	dt := f()
	v := xerror.PanicBytes(jsonx.Marshal(dt))
	return byteutil.ToStr(v)
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
