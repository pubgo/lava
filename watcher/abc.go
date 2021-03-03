package watcher

import (
	"encoding/json"

	"github.com/pubgo/xerror"
)

type Option func(opts *Options)
type Options struct {
	prefix bool
}

// Watcher ...
type Watcher interface {
	Name() string
	Start() error
	Close() error
}

type CallBack func(event *Response) error
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
	defer xerror.Resp(func(err xerror.XErr) {
		gErr = err.WrapF("input: %#v, output: %t", t.Value, val)
	})

	return xerror.Wrap(json.Unmarshal(t.Value, val))
}

func (t *Response) checkEventType() error {
	switch t.Event {
	case "DELETE", "PUT":
		return nil
	default:
		return xerror.New("unknown type")
	}
}
