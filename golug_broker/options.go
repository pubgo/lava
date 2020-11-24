package golug_broker

import (
	"context"
	"crypto/tls"
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

// Addrs sets the host addresses to be used by the broker
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// Queue sets the name of the queue to share messages on
func Queue(name string) SubOption {
	return func(o *SubOptions) {
		o.Queue = name
	}
}

// Secure communication with the broker
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// SubscribeContext set context
func SubscribeContext(ctx context.Context) SubOption {
	return func(o *SubOptions) {
		o.Ctx = ctx
	}
}
