package plugin

import (
	"github.com/pubgo/xerror"
)

const defaultModule = "__global"

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
	defer xerror.RespRaise(func(err xerror.XErr) error { return err.Wrap("register plugin error") })

	xerror.Assert(pg == nil, "plugin[pg] is nil")

	name := pg.String()
	xerror.Assert(name == "", "plugin name is null")

	opt := options{Module: defaultModule}
	for _, o := range opts {
		o(&opt)
	}

	pgs := plugins[opt.Module]
	for i := range pgs {
		xerror.Assert(pgs[i].String() == name, "plugin [%s] already registers", name)
	}
	plugins[opt.Module] = append(plugins[opt.Module], pg)
}
