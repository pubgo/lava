package watcher

import (
	"context"

	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

type CallBack func(name string, event *Response) error

// Watcher ...
type Watcher interface {
	Get(ctx context.Context, opts ...Opt) ([]*Response, error)
	GetCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt) error
	WatchCallback(ctx context.Context, key string, fn func(resp *Response), opts ...Opt)
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
	Close(opts ...Opt)
	Name() string
}

type Opt func(*Opts)
type Opts struct{}

type Response struct {
	Event    string
	Key      string
	Value    []byte
	Revision int64
}

func (t *Response) OnPut(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == "PUT" {
		fn()
	}
}

func (t *Response) OnDelete(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == "DELETE" {
		fn()
	}
}

func (t *Response) Decode(val interface{}) error {
	return xerror.WrapF(jsonx.Unmarshal(t.Value, val), "input: %s, output: %#v", t.Value, val)
}

func (t *Response) checkEventType() error {
	switch t.Event {
	case "DELETE", "PUT":
		return nil
	default:
		return xerror.Fmt("unknown type: %s", t.Event)
	}
}
