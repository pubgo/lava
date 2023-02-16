package migrates

import (
	"github.com/pubgo/lava/core/migrates"
)

func Migrations() []migrates.Migrate {
	return []migrates.Migrate{
		m0001,
	}
}
