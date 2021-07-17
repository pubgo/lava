package task

import (
	"github.com/pubgo/lug/broker"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
)

type Handler = broker.Handler
type Opts = broker.SubOpts
type Entry interface {
	entry.Entry
	Plugin(plugins ...plugin.Plugin)
	Register(topic string, handler Handler, opts ...*Opts)
}
