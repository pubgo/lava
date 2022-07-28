package casbin

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
)

func init() {
	defer recovery.Exit()
	dix.Provider(New)
}
