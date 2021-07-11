package registry

import (
	"github.com/pubgo/xerror"
)

var defaultRegistry Registry

func Default() Registry { return defaultRegistry }

func Init() {
	var cfg = GetDefaultCfg()
	defaultRegistry = xerror.ExitErr(cfg.Build()).(Registry)
}
