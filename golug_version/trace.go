package golug_version

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch("golug_version", func() interface{} { return List() })
}
