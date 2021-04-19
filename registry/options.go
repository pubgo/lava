package registry

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/pubgo/xerror"
)

func TTL(t string) RegOpt {
	return func(o *RegOpts) {
		dur, err := time.ParseDuration(t)
		xerror.Panic(err)
		o.TTL = dur
	}
}

type Opts struct {
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
	TLSConfig *tls.Config
	Context   context.Context
}

type RegOpts struct {
	TTL     time.Duration
	Context context.Context
}

type WatchOpts struct {
	Service string
	Context context.Context
}

type DeRegOpts struct {
	Context context.Context
}

type GetOpts struct {
	Timeout time.Duration
}

type ListOpts struct {
	Context context.Context
}

// Addrs is the registry addresses to use
func Addrs(addrs ...string) Opt {
	return func(o *Opts) {
		o.Addrs = addrs
	}
}

func Timeout(t time.Duration) Opt {
	return func(o *Opts) {
		o.Timeout = t
	}
}

// Secure communication with the registry
func Secure(b bool) Opt {
	return func(o *Opts) {
		o.Secure = b
	}
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Opt {
	return func(o *Opts) {
		o.TLSConfig = t
	}
}

func RegisterTTL(t time.Duration) RegOpt {
	return func(o *RegOpts) {
		o.TTL = t
	}
}

func RegisterContext(ctx context.Context) RegOpt {
	return func(o *RegOpts) {
		o.Context = ctx
	}
}

// Watch a service
func WatchService(name string) WatchOpt {
	return func(o *WatchOpts) {
		o.Service = name
	}
}

func WatchContext(ctx context.Context) WatchOpt {
	return func(o *WatchOpts) {
		o.Context = ctx
	}
}

func DeRegisterContext(ctx context.Context) DeRegOpt {
	return func(o *DeRegOpts) {
		o.Context = ctx
	}
}

func ListContext(ctx context.Context) ListOpt {
	return func(o *ListOpts) {
		o.Context = ctx
	}
}

type servicesKey struct{}

// Services is an option that preloads service data
func Services(s map[string][]*Service) Opt {
	return func(o *Opts) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, servicesKey{}, s)
	}
}
