package tracing

import (
	"github.com/pubgo/lava/types"
)

func init() {
	RegisterFactory("noop", func(cfg types.CfgMap) error { return nil })
}
