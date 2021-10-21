package tracing

import "github.com/pubgo/xerror"

func init() {
	xerror.Exit(Register("noop", func(cfg map[string]interface{}) error { return nil }))
}
