package golug_env

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch("env", func() interface{} { return List() })
}
