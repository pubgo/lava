package tracing

import (
	"github.com/pubgo/lava/config"
)

func init() {
	RegisterFactory("noop", func(cfg config.CfgMap) error { return nil })
}
