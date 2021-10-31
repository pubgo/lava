package plugin

import (
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

var plugins = make(map[string]Plugin)

func All() map[string]Plugin { return plugins }
func Get(name string) Plugin { return plugins[name] }

func Middleware(name string, middleware types.Middleware) {
	Register(&Base{Name: name, OnMiddleware: middleware})
}

func Register(pg Plugin) {
	defer xerror.RespExit("register plugin error")
	xerror.Assert(pg == nil, "plugin[pg] is nil")
	xerror.Assert(pg.UniqueName() == "", "plugin name is null")
	xerror.Assert(plugins[pg.UniqueName()] != nil, "plugin [%s] already exists", pg.UniqueName())
	plugins[pg.UniqueName()] = pg
}
