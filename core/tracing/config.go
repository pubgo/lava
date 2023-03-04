package tracing

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
)

const Name = "tracing"

type Cfg struct {
	Driver    string                 `json:"driver"`
	DriverCfg map[string]interface{} `json:"driver_config"`
}

func (cfg Cfg) Build() (err error) {
	defer recovery.Err(&err, func(err *errors.Event) {
		err.Any("cfg", cfg)
	})

	driver := cfg.Driver
	assert.If(driver == "", "tracer driver is null")

	fc := GetFactory(driver)
	assert.If(fc == nil, "tracer driver [%s] not found", driver)

	return fc(cfg.DriverCfg)
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Driver: "noop",
	}
}
