package watcher

import (
	"context"

	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

type Factory func(cfg map[string]interface{}) (Watcher, error)
type CallBack func(name string, event *Response) error

// Watcher ...
type Watcher interface {
	Watch(ctx context.Context, key string, opts ...Opt) <-chan *Response
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

func (t *Response) Decode(val interface{}) (gErr error) {
	defer xerror.RespErr(&gErr)

	var err = jsonx.Unmarshal(t.Value, val)
	return xerror.WrapF(err, "input: %s, output: %#v", t.Value, val)
}

func (t *Response) checkEventType() error {
	switch t.Event {
	case "DELETE", "PUT":
		return nil
	default:
		return xerror.New("unknown type")
	}
}
