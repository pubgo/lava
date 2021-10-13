package async

import (
	"reflect"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/syncx"
)

func Async(fn func(ctx Ctx) (interface{}, error)) Value {
	value := newValue()
	if fn == nil {
		value.complete(nil, xerror.Fmt("[fn] should not be nil"))
		return value
	}

	syncx.SafeGo(func() {
		defer xerror.Resp(func(err xerror.XErr) {
			value.complete(nil, err.WrapF("recovery error, func:%s, caller:%s", reflect.TypeOf(fn), stack.Func(fn)))
		})

		value.complete(fn(value.context()))
	})

	return value
}

func Pipe(val Value, fn func(data interface{}) (interface{}, error)) Value {
	var value = newValue()

	if val == nil {
		value.complete(nil, xerror.Fmt("[val] should not be nil"))
		return value
	}

	if fn == nil {
		value.complete(nil, xerror.Fmt("[fn] should not be nil"))
		return value
	}

	value.context().addCancel(val.Cancel)

	syncx.SafeGo(func() {
		if err := val.Err(); err != nil {
			value.complete(nil, err)
			return
		}

		defer xerror.Resp(func(err xerror.XErr) {
			value.complete(nil, err.WrapF("recovery error, func:%s, caller:%s", reflect.TypeOf(fn), stack.Func(fn)))
		})

		value.complete(fn(val.Get()))
	})

	return value
}
