package golug_log

import (
	"github.com/pubgo/golug/vars"
)

func init() {
	vars.Watch(name, func() interface{} { return cfg })
}
