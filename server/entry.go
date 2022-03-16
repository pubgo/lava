package server

import (
	"github.com/pubgo/lava/types"
	"github.com/urfave/cli/v2"
)

type Runtime interface {
	InitRT()
	Start() error
	Stop() error
	Options() Opts
	MiddlewareInter(middleware types.Middleware)
}

type Entry interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	Middleware(middleware types.Middleware)
	Description(description ...string)
	Flags(flags cli.Flag)
	Commands(commands *cli.Command)
}

type Handler interface {
	Init()
}

// AssertHandler handler校验
func AssertHandler(Handler) error { return nil }