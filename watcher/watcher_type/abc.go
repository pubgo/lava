package watcher_type

import (
	"context"

	"github.com/pubgo/lava/config/config_type"
)

const Name = "watcher"

type (
	Response = WatchResp
	Opt      func(*options)
	options  struct{}
	Factory  = func(cfg config_type.CfgMap) (Watcher, error)
)

// Watcher ...
type Watcher interface {
	Init()
	Close(ctx context.Context, opts ...Opt)
	Get(ctx context.Context, key string, opts ...Opt) ([]*Response, error)
	GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
	WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt)
}