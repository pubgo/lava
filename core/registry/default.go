package registry

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/core/registry/registry_type"
	"github.com/pubgo/xerror"
)

var DefaultRegistry registry_type.Registry

func Default() registry_type.Registry { return DefaultRegistry }

func Init(driver string, cfg config_type.CfgMap) (err error) {
	xerror.RespErr(&err)

	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	xerror.Assert(driver == "", "registry driver is null")
	xerror.Assert(!factories.Has(driver), "registry driver %s not found", driver)

	var fc = factories.Get(driver).(Factory)
	DefaultRegistry = xerror.PanicErr(fc(cfg)).(registry_type.Registry)
	return nil
}
