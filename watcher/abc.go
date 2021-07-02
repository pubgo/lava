package watcher

import (
	"context"

	"github.com/pubgo/lug/types"
)

type Response = types.Response

type CallBack func(name string, resp *Response) error

// Watcher ...
type Watcher interface {
	Name() string
	Close(ctx context.Context, opts ...Opt)
	Get(ctx context.Context, key string, opts ...Opt) ([]*Response, error)
	GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
	WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt)
}

type Opt func(*options)
type options struct{}
