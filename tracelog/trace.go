package tracelog

import (
	"expvar"
	"github.com/pubgo/xerror"
)

func Watch(name string, data func() interface{}) {
	expvar.Publish(name, expvar.Func(func() (val interface{}) {
		defer xerror.Resp(func(err xerror.XErr) { val = err })
		return data()
	}))
}
