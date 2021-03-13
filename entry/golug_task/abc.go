package golug_task

import (
	"github.com/pubgo/golug/broker"
	"github.com/pubgo/golug/entry"
)

type Entry interface {
	entry.Entry
	Register(topic string, handler broker.Handler, opts ... *broker.SubOpts)
}
