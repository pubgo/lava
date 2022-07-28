package encoding

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/vars"
)

func init() {
	defer recovery.Exit()
	vars.Register(Name, func() interface{} { return Keys() })
}
