package registry

import (
	"github.com/pubgo/xerror"
)

var defaultRegistry Registry

func Default() Registry { return defaultRegistry }

func Init(driver string, cfg map[string]interface{}) (err error) {
	xerror.RespErr(&err)

	if cfg == nil {
		cfg = make(map[string]interface{})
	}

	xerror.Assert(driver == "", "registry driver is null")
	xerror.Assert(!factories.Has(driver), "registry driver %s not found", driver)

	var fc = factories.Get(driver).(Factory)
	defaultRegistry = xerror.PanicErr(fc(cfg)).(Registry)
	return nil
}
