package log_plugin

import (
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch(name, func() interface{} { return cfg })
}
