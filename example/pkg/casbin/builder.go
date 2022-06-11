package casbin

import "github.com/pubgo/dix"

func init() {
	dix.Register(New)
}
