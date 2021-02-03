package golug_log

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch(Name, func() interface{} { return cfg })
}
