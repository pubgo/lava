package encoding

import "github.com/pubgo/funk/vars"

func init() {
	vars.Register(Name, func() interface{} { return Keys() })
}
