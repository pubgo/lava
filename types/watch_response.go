package types

import (
	"github.com/pubgo/xerror"
)

type WatchResp struct {
	Type    string
	Event   EventType
	Key     string
	Value   []byte
	Version int64
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
