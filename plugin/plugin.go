package plugin

import (
	"github.com/pubgo/xerror"
)

const defaultModule = "__global"

var plugins = make(map[string]map[string]Plugin)

func All() map[string]map[string]Plugin { return plugins }

func List(opts ...Opt) map[string]Plugin {
	mOpts := options{Module: defaultModule}
	for _, o := range opts {
		o(&mOpts)
	}
	return plugins[mOpts.Module]
}

func ListWithDefault(opts ...Opt) map[string]Plugin {
	mOpts := options{Module: defaultModule}
	for _, o := range opts {
		o(&mOpts)
	}

	var plgList = plugins[defaultModule]
	for k, v := range plugins[mOpts.Module] {
		plgList[k] = v
	}
	return plgList
}

func Get(name string, opts ...Opt) Plugin {
	mOpts := options{Module: defaultModule}
	for _, o := range opts {
		o(&mOpts)
	}
	return plugins[mOpts.Module][name]
}

func Register(pg Plugin, opts ...Opt) {
	defer xerror.RespExit("register plugin error")

	xerror.Assert(pg == nil, "plugin[pg] is nil")
	xerror.Assert(pg.Id() == "", "plugin name is null")

	opt := options{Module: defaultModule}
	for _, o := range opts {
		o(&opt)
	}

	if plugins[opt.Module] == nil {
		plugins[opt.Module] = make(map[string]Plugin)
	}

	pgs := plugins[opt.Module]
	xerror.Assert(pgs[pg.Id()] != nil, "plugin [%s] already exists", pg.Id())

	plugins[opt.Module][pg.Id()] = pg
}
