package golug_watcher

import (
	"encoding/json"

	"github.com/pubgo/xerror"
)

// Watcher ...
type Watcher interface {
	String() string
	Start() error
	Close() error
	List() []string
	Watch(name string, h CallBack) (err error)
	Remove(name string) (err error)
}

type CallBack func(event *Response) error
type Response struct {
	Event    string
	Key      string
	Value    []byte
	Revision int64
}

func (t *Response) Decode(val interface{}) (err error) {
	defer xerror.RespErr(&err)
	return xerror.Wrap(json.Unmarshal(t.Value, val))
}
