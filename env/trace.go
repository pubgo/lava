package env

import (
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch("env", func() interface{} { return List() })
}
