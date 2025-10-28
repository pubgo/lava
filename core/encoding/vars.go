package encoding

import "github.com/pubgo/funk/v2/vars"

func init() {
	vars.Register(Name, func() any { return Keys() })
	vars.Register(Name+"-mapping", func() any { return cdcMapping })
}
