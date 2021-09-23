package task

import (
	broker2 "github.com/pubgo/lug/abc/broker"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
)

type Handler = broker2.Handler
type Opts = broker2.SubOpts
type Entry interface {
	entry.Entry
	Plugin(plugins ...plugin.Plugin)
	Register(topic string, handler Handler, opts ...*Opts)
}
