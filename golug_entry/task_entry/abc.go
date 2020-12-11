package task_entry

import (
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_entry"
)

type Entry interface {
	golug_entry.Entry
	Register(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) error
}
