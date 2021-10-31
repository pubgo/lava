package task

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/plugins/broker"
)

type Handler = broker.Handler
type Opts = broker.SubOpts
type Entry interface {
	entry.Entry
	Register(topic string, handler Handler, opts ...*Opts)
}
