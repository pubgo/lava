package vars

import (
	"expvar"

	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

func Float(name string) *expvar.Float {
	var v = expvar.Get(name)
	if v == nil {
		return expvar.NewFloat(name)
	}
	return v.(*expvar.Float)
}

func Int(name string) *expvar.Int {
	var v = expvar.Get(name)
	if v == nil {
		return expvar.NewInt(name)
	}
	return v.(*expvar.Int)
}

func String(name string) *expvar.String {
	var v = expvar.Get(name)
	if v == nil {
		return expvar.NewString(name)
	}
	return v.(*expvar.String)
}

func Map(name string) *expvar.Map {
	var v = expvar.Get(name)
	if v == nil {
		return expvar.NewMap(name)
	}
	return v.(*expvar.Map)
}

type value func() interface{}

func (f value) Value() interface{} { return f() }
func (f value) String() (r string) {
	defer xerror.Resp(func(err xerror.XErr) { r = err.Stack() })

	dt := f()

	if _, ok := dt.(string); ok {
		return dt.(string)
	}

	v := xerror.PanicBytes(jsonx.Marshal(dt))
	return byteutil.ToStr(v)
}

func Register(name string, data func() interface{}) {
	expvar.Publish(name, value(data))
}

func Has(name string) bool {
	return expvar.Get(name) != nil
}

func Each(fn func(key string, val expvar.Var)) {
	expvar.Do(func(kv expvar.KeyValue) { fn(kv.Key, kv.Value) })
}
