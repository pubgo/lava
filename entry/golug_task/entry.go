package golug_task

import (
	"github.com/pubgo/golug/broker"
	"github.com/pubgo/golug/entry/base"
	"github.com/pubgo/xerror"
)

var _ Entry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler broker.Handler
	opts    broker.SubOptions
	optList []broker.SubOption
}

type taskEntry struct {
	*base.Entry
	cfg      Cfg
	broker   broker.Broker
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler broker.Handler, opts ...broker.SubOption) error {
	var opts1 broker.SubOptions
	for i := range opts {
		opts[i](&opts1)
	}
	opts1.Topic = topic

	t.handlers = append(t.handlers, entryTaskHandler{handler: handler, opts: opts1, optList: opts})
	return nil
}

func (t *taskEntry) Stop() (err error) {
	return nil
}

func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	t.broker = broker.Get(t.cfg.Broker)
	for i := range t.handlers {
		handler := t.handlers[i]
		brk := t.broker
		if handler.opts.Broker != nil {
			brk = handler.opts.Broker
		}

		xerror.Panic(brk.Subscribe(handler.opts.Topic, handler.handler, handler.optList...))
	}

	return nil
}

func newEntry(name string) *taskEntry {
	ent := &taskEntry{Entry: base.New(name)}
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent

}
func New(name string) Entry { return newEntry(name) }
