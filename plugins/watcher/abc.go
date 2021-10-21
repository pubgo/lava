package watcher

import (
	"context"

	"github.com/pubgo/lava/types"
)

type (
	Response = types.WatchResp
	CallBack func(name string, resp *Response) error
	Opt      func(*options)
	options  struct{}
)

// Watcher ...
type Watcher interface {
	Name() string
	Close(ctx context.Context, opts ...Opt)
	Get(ctx context.Context, key string, opts ...Opt) ([]*Response, error)
	GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
	WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt)
}
