package golug_entry

import (
	"github.com/pubgo/golug/golug_broker"
)

type TaskEntry interface {
	Entry
	Register(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) error
}
