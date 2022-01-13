package tracing

import (
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(RegisterFactory("noop", func(cfg types.CfgMap) error { return nil }))
}
