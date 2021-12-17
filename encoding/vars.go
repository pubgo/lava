package encoding

import (
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register(Name, func() interface{} { return Keys() })
}
