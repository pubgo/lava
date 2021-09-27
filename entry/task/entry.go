package task

import (
	"github.com/pubgo/lug/abc/broker"
	"github.com/pubgo/lug/entry/base"

	"github.com/pubgo/xerror"
)

var _ Entry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler broker.Handler
	opts    *broker.SubOpts
}

type taskEntry struct {
	*base.Entry
	cfg      Cfg
	broker   broker.Broker
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler Handler, opts ...*Opts) {
	var taskHandler = entryTaskHandler{handler: handler, opts: new(broker.SubOpts)}
	if len(opts) > 0 && opts[0] != nil {
		taskHandler.opts = opts[0]
	}

	taskHandler.opts.Topic = topic
	t.handlers = append(t.handlers, taskHandler)
}

func (t *taskEntry) Stop() (err error) { return nil }
func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	t.broker = broker.Get(t.cfg.Broker)
	for i := range t.handlers {
		handler := t.handlers[i]
		brk := t.broker
		if handler.opts.Broker != nil {
			brk = handler.opts.Broker
		}

		xerror.Panic(brk.Sub(handler.opts.Topic, handler.handler, handler.opts))
	}

	return nil
}

func newEntry(name string) *taskEntry {
	return &taskEntry{Entry: base.New(name)}
}

func New(name string) Entry { return newEntry(name) }
