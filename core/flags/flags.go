package flags

import (
	"github.com/pubgo/redant"
)

var flags []redant.Option

func Register(flag redant.Option) {
	flags = append(flags, flag)
}

func GetFlags() []redant.Option { return flags }
