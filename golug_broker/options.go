package golug_broker

import (
	"context"
)

// WithPubCtx set context
func WithPubCtx(ctx context.Context) PubOption {
	return func(o *PubOptions) {
		o.Context = ctx
	}
}

func NewSubOptions(opts ...SubOption) SubOptions {
	opt := SubOptions{}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// WithQueue sets the name of the queue to share messages on
func WithQueue(name string) SubOption {
	return func(o *SubOptions) {
		o.Queue = name
	}
}

// WithSubCtx set context
func WithSubCtx(ctx context.Context) SubOption {
	return func(o *SubOptions) {
		o.Ctx = ctx
	}
}
