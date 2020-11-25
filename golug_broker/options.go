package golug_broker

import (
	"context"
)

// PublishContext set context
func PublishContext(ctx context.Context) PubOption {
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

// Queue sets the name of the queue to share messages on
func Queue(name string) SubOption {
	return func(o *SubOptions) {
		o.Queue = name
	}
}

// SubscribeContext set context
func SubscribeContext(ctx context.Context) SubOption {
	return func(o *SubOptions) {
		o.Ctx = ctx
	}
}
