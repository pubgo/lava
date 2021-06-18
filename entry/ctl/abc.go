package ctl

import (
	"context"

	"github.com/pubgo/lug/entry"
)

type Entry interface {
	entry.Entry
	Register(fn func(ctx context.Context), optList ...Opt)
	RegisterLoop(fn func(ctx context.Context), optList ...Opt)
}

type Opt func(opts *options)
type options struct {
	Name    string
	once    bool
	handler func(ctx context.Context)
	cancel  context.CancelFunc
}

func WithName(name string) Opt {
	return func(opts *options) {
		opts.Name = name
	}
}
