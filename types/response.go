package types

import (
	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
	"gopkg.in/yaml.v2"
)

const (
	PUT    = "PUT"
	DELETE = "DELETE"
)

type Response struct {
	Type    string
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

func unmarshal(in []byte, c interface{}) (err error) {
	defer func() {
		if err != nil {
			err = xerror.Fmt("Unmarshal Error, encoding\n")
		}
	}()

	// "yaml", "yml"
	if err = yaml.Unmarshal(in, &c); err == nil {
		return
	}

	// "json"
	if err = jsonx.Unmarshal(in, &c); err == nil {
		return
	}

	// "hcl"
	if err = hcl.Unmarshal(in, &c); err == nil {
		return
	}

	// "toml"
	if err = toml.Unmarshal(in, &c); err == nil {
		return
	}

	return
}
