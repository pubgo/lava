package vars

import (
	"encoding/json"
	"expvar"
	"fmt"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/jsonx"

	"github.com/pubgo/lava/pkg/utils"
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
	defer recovery.Recovery(func(err xerr.XErr) { r = err.String() })

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
	case json.Marshaler:
		ret := result.Wrap(jsonx.Marshal(dt))
		if ret.IsErr() {
			return xerr.WrapXErr(ret.Err().Err()).Stack()
		}
		return utils.BtoS(ret.Unwrap())
	}
	return fmt.Sprintf("%v", dt)
}

func Register(name string, data func() interface{}) {
	expvar.Publish(name, Value(data))
}

func Has(name string) bool {
	return expvar.Get(name) != nil
}

func Each(fn func(key string, val expvar.Var)) {
	expvar.Do(func(kv expvar.KeyValue) { fn(kv.Key, kv.Value) })
}
