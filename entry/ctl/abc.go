package ctl

import (
	"context"

	"github.com/pubgo/x/fx"

	"github.com/pubgo/lava/entry"
)

type Handler func(fx.Ctx)
type Service interface {
	Run() map[string]Handler
	RunLoop() map[string]Handler
}

type Entry interface {
	entry.Entry
	Register(run Service, optList ...Opt)
}

type Opt func(opts *options)
type options struct {
	Name    string
	once    bool
	handler Handler
	cancel  context.CancelFunc
}

func withName(name string) Opt {
	return func(opts *options) {
		opts.Name = name
	}
}
