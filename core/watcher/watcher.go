package watcher

import (
	"github.com/goccy/go-json"
	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/xerror"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/event"
)

type Response struct {
	Type    string
	Event   event.EventType
	Key     string
	Value   []byte
	Version int64
}

func (t *Response) Decode(c interface{}) (err error) {
	defer xerror.RespErr(&err)

	var data = t.Value

	// "yaml", "yml"
	if err = yaml.Unmarshal(data, &c); err == nil {
		return
	}

	// "json"
	if err = json.Unmarshal(data, &c); err == nil {
		return
	}

	// "toml"
	if err = toml.Unmarshal(data, &c); err == nil {
		return
	}

	// "hcl"
	if err = hcl.Unmarshal(data, &c); err == nil {
		return
	}

	return errors.Unknown("watcher.decode", "data=>%s, c=>%T", data, c)
}

func (t *Response) OnPut(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == event.EventType_UPDATE || t.Event == event.EventType_CREATE {
		fn()
	}
}

func (t *Response) OnDelete(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == event.EventType_DELETE {
		fn()
	}
}

func (t *Response) checkEventType() error {
	switch t.Event {
	case event.EventType_UPDATE, event.EventType_DELETE, event.EventType_CREATE:
		return nil
	default:
		return xerror.Fmt("unknown event: %s", t.Event)
	}
}
