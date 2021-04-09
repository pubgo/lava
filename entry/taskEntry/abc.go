package taskEntry

import (
	"github.com/pubgo/lug/broker"
	"github.com/pubgo/lug/entry"
)

type Entry interface {
	entry.Entry
	Register(topic string, handler broker.Handler, opts ... *broker.SubOpts)
}
