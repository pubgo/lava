package golug_watcher

import (
	"github.com/pubgo/golug/golug_consts"
)

const (
	PUT    Event = 0
	DELETE Event = 1
)

type Event int32

func (t Event) String() string {
	switch t {
	case 0:
		return "PUT"
	case 1:
		return "DELETE"
	default:
		return golug_consts.Unknown
	}
}
