package env

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch("env", func() interface{} { return List() })
}
