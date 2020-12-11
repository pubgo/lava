package golug_broker

import (
	"context"
)

// WithPublishCtx set context
func WithPublishCtx(ctx context.Context) PubOption {
	return func(o *PubOptions) {
		o.Context = ctx
	}
}

func NewSubscribeOptions(opts ...SubOption) SubOptions {
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

// WithSubscribeCtx set context
func WithSubscribeCtx(ctx context.Context) SubOption {
	return func(o *SubOptions) {
		o.Ctx = ctx
	}
}
