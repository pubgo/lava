package ctl

import (
	"context"

	"github.com/pubgo/x/fx"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
)

type Entry interface {
	entry.Entry
	Plugin(plugins ...plugin.Plugin)
	Register(name string, fn func(ctx fx.Ctx), optList ...Opt)
	RegisterLoop(name string, fn func(ctx fx.Ctx), optList ...Opt)
}

type Opt func(opts *options)
type options struct {
	Name    string
	once    bool
	handler func(ctx fx.Ctx)
	cancel  context.CancelFunc
}

func withName(name string) Opt {
	return func(opts *options) {
		opts.Name = name
	}
}
