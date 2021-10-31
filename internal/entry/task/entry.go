package task

import (
	broker3 "github.com/pubgo/lava/plugins/broker"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/entry/base"
)

func newEntry(name string) *taskEntry {
	return &taskEntry{Entry: base.New(name)}
}

func New(name string) Entry { return newEntry(name) }

var _ Entry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler broker3.Handler
	opts    *broker3.SubOpts
}

type taskEntry struct {
	*base.Entry
	cfg      Cfg
	broker   broker3.Broker
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler Handler, opts ...*Opts) {
	var taskHandler = entryTaskHandler{handler: handler, opts: new(broker3.SubOpts)}
	if len(opts) > 0 && opts[0] != nil {
		taskHandler.opts = opts[0]
	}

	taskHandler.opts.Topic = topic
	t.handlers = append(t.handlers, taskHandler)
}

func (t *taskEntry) Stop() (err error) { return nil }
func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	t.broker = broker3.Get(t.cfg.Broker)
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
