package encoding

import (
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Watch(Name, func() interface{} { return Keys() })
}
