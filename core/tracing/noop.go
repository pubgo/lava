package tracing

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/config"
)

func init() {
	defer recovery.Exit()

	RegisterFactory("noop", func(cfg config.CfgMap) error { return nil })
}
