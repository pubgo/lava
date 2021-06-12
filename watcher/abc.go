package watcher

import (
	"context"

	"github.com/pubgo/xerror"
)

const (
	PUT    = "PUT"
	DELETE = "DELETE"
)

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

type Response struct {
	Event   string
	Key     string
	Value   []byte
	Version int64
}

func (t *Response) OnPut(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == PUT {
		fn()
	}
}

func (t *Response) OnDelete(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == DELETE {
		fn()
	}
}

func (t *Response) Decode(val interface{}) error {
	return xerror.WrapF(unmarshal(t.Value, val), "input: %s, output: %#v", t.Value, val)
}

func (t *Response) checkEventType() error {
	switch t.Event {
	case DELETE, PUT:
		return nil
	default:
		return xerror.Fmt("unknown event: %s", t.Event)
	}
}
