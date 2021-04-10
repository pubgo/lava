package task

import (
	"github.com/pubgo/lug/broker"
	"github.com/pubgo/lug/entry"
)

type Handler = broker.Handler
type Opts = broker.SubOpts
type Entry interface {
	entry.Abc
	Register(topic string, handler Handler, opts ...*Opts)
}
