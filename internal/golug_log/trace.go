package golug_log

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch(name, func() interface{} { return cfg })
}
