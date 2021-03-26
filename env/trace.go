package env

import (
	"github.com/pubgo/golug/vars"
)

func init() {
	vars.Watch("env", func() interface{} { return List() })
}
