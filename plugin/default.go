package plugin

import (
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"
)

const defaultModule = "__default"

var plugins = make(map[string][]Plugin)

func All() map[string][]Plugin {
	pls := make(map[string][]Plugin, len(plugins))
	for k, v := range plugins {
		pls[k] = append(pls[k], v...)
	}
	return pls
}

func List(opts ...Opt) []Plugin {
	mOpts := options{Module: defaultModule}
	for _, o := range opts {
		o(&mOpts)
	}

	return plugins[mOpts.Module]
}

func Register(pg Plugin, opts ...Opt) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Fatal("register plugin", zap.Any("err", err))
	})

	if pg == nil {
		xlog.Fatal("plugin[pg] is nil")
		return
	}

	name := pg.String()
	xerror.Assert(name == "", "plugin name is null")

	options := options{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	pgs := plugins[options.Module]
	for i := range pgs {
		xerror.Assert(pgs[i].String() == name, "plugin [%s] already registers", name)
	}
	plugins[options.Module] = append(plugins[options.Module], pg)
}
