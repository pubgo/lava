package vars

import (
	"expvar"

	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

type value func() interface{}

func (f value) Value() interface{} { return f() }
func (f value) String() string {
	v, err := jsonx.Marshal(f())
	if err != nil {
		return err.Error()
	}

	return byteutil.ToStr(v)
}

func Watch(name string, data func() interface{}) {
	expvar.Publish(name, value(func() (val interface{}) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("unknown error",
				xlog.String("pkg", "expvar"),
				xlog.String("name", name),
				xlog.Any("err", err))
		})
		return data()
	}))
}
