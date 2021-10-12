package task

import (
	"github.com/pubgo/lava/abc/broker"
	"github.com/pubgo/lava/entry"
)

type Handler = broker.Handler
type Opts = broker.SubOpts
type Entry interface {
	entry.Entry
	Register(topic string, handler Handler, opts ...*Opts)
}
