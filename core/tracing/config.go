package tracing

import ()

const Name = "tracing"

type Cfg struct {
	Driver    string                 `json:"driver"`
	DriverCfg map[string]interface{} `json:"driver_config"`
}

func (cfg Cfg) Build() (err error) {
	defer xerror.RecoverErr(&err, func(err xerror.XErr) xerror.XErr {
		return err.WrapF("cfg=>\n\t%#v", cfg)
	})

	driver := cfg.Driver
	xerror.Assert(driver == "", "tracer driver is null")

	fc := GetFactory(driver)
	xerror.Assert(fc == nil, "tracer driver [%s] not found", driver)

	return fc(cfg.DriverCfg)
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Driver: "noop",
	}
}
