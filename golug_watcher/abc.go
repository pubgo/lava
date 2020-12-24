package golug_watcher

import (
	"encoding/json"

	"github.com/pubgo/xerror"
)

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

func (t *Response) Decode(val interface{}) (err error) {
	return xerror.Try(func() { xerror.Panic(json.Unmarshal(t.Value, val)) })
}
