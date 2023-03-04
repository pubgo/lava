package vars

import (
	"expvar"
	"fmt"

	jjson "github.com/goccy/go-json"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
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

type Value func() interface{}

func (f Value) Value() interface{} { return f() }

func (f Value) String() (r string) {
	defer recovery.Recovery(func(err error) {
		ret := result.Wrap(jjson.Marshal(err))
		if ret.IsErr() {
			r = pretty.Sprint(ret.Err())
		} else {
			r = convert.B2S(ret.Unwrap())
		}
	})

	dt := f()
	switch dt.(type) {
	case nil:
		return "null"
	case string:
		return dt.(string)
	case []byte:
		return string(dt.([]byte))
	case fmt.Stringer:
		return dt.(fmt.Stringer).String()
	default:
		ret := result.Wrap(jjson.Marshal(dt))
		if ret.IsErr() {
			return pretty.Sprint(ret.Err())
		}
		return convert.B2S(ret.Unwrap())
	}
}

func Register(name string, data func() interface{}) {
	defer recovery.Exit()
	assert.If(Has(name), "name:%s already exists", name)
	expvar.Publish(name, Value(data))
}

func Has(name string) bool {
	return expvar.Get(name) != nil
}

func Each(fn func(key string, val expvar.Var)) {
	expvar.Do(func(kv expvar.KeyValue) { fn(kv.Key, kv.Value) })
}
