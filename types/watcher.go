package types

import (
	"github.com/goccy/go-json"
	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/xerror"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lava/errors"
)

type Watcher func(name string, r *WatchResp) error

type WatchResp struct {
	Type    string
	Event   EventType
	Key     string
	Value   []byte
	Version int64
}

// Decode ...
func (t *WatchResp) Decode(c interface{}) error {
	return Decode(t.Value, c)
}

func (t *WatchResp) OnPut(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == EventType_UPDATE {
		fn()
	}
}

func (t *WatchResp) OnDelete(fn func()) {
	xerror.Panic(t.checkEventType())
	if t.Event == EventType_DELETE {
		fn()
	}
}

func (t *WatchResp) checkEventType() error {
	switch t.Event {
	case EventType_UPDATE, EventType_DELETE:
		return nil
	default:
		return xerror.Fmt("unknown event: %s", t.Event)
	}
}

func Decode(data []byte, c interface{}) (err error) {
	defer xerror.RespErr(&err)

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

	return errors.Unknown("config.watcher.decode", "data=>%s, c=>%T", data, c)
}
