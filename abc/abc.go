package abc

import (
	"github.com/urfave/cli/v2"
)

type Init interface {
	Init()
}

type Close interface {
	Close()
}

type Flags interface {
	Flags() []cli.Flag
}
