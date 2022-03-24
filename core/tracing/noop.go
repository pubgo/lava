package tracing

import (
	"github.com/pubgo/lava/config/config_type"
)

func init() {
	RegisterFactory("noop", func(cfg config_type.CfgMap) error { return nil })
}
