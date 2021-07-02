package plugin

import (
	"github.com/pubgo/lug/pkg/logutil"
	"github.com/pubgo/xerror"
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
	logutil.Exit(func() {
		xerror.Assert(pg == nil || pg.String() == "", "plugin is nil")

		options := options{Module: defaultModule}
		for _, o := range opts {
			o(&options)
		}

		name := pg.String()
		xerror.Assert(name == "", "plugin name is null")

		pgs := plugins[options.Module]
		for i := range pgs {
			xerror.Assert(pgs[i].String() == name, "plugin [%s] already registers", name)
		}
		plugins[options.Module] = append(plugins[options.Module], pg)
	})
}
