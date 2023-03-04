package encoding

import "github.com/pubgo/lava/core/vars"

func init() {
	vars.Register(Name, func() interface{} { return Keys() })
}
