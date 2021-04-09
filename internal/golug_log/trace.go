package golug_log

import (
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch(name, func() interface{} { return cfg })
}
