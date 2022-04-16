package watcher

import (
	"context"

	"github.com/pubgo/lava/config"
)

const Name = "watcher"

type (
	Opt          func(*options)
	options      struct{}
	Factory      = func(cfg config.CfgMap) (Watcher, error)
	WatchHandler = func(name string, r *Response) error
)

// Watcher ...
type Watcher interface {
	Init()
	Close()
	Get(ctx context.Context, key string, opts ...Opt) ([]*Response, error)
	GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
	WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt)
}