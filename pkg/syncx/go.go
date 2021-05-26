package syncx

import (
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func Next(name string, cb func()) {
	defer xerror.Raise(func(err xerror.XErr) error {
		return err.WrapF("name:%s, stack:%s", name, stack.Func(cb))
	})

	cb()
}

func Log(name string, cb func()) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error(name, xlog.Any("err", err))
	})

	cb()
}
