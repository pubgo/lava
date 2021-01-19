package golug_task

import (
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/xerror"
)

var _ Entry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler golug_broker.Handler
	opts    golug_broker.SubOptions
	optList []golug_broker.SubOption
}

type taskEntry struct {
	*golug_base.Entry
	cfg      Cfg
	broker   golug_broker.Broker
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) error {
	var opts1 golug_broker.SubOptions
	for i := range opts {
		opts[i](&opts1)
	}
	opts1.Topic = topic

	t.handlers = append(t.handlers, entryTaskHandler{handler: handler, opts: opts1, optList: opts})
	return nil
}

func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	t.broker = golug_broker.Get(t.cfg.Broker)
	for i := range t.handlers {
		handler := t.handlers[i]
		broker := t.broker
		if handler.opts.Broker != nil {
			broker = handler.opts.Broker
		}

		xerror.Panic(broker.Subscribe(handler.opts.Topic, handler.handler, handler.optList...))
	}

	return nil
}

func newEntry(name string) *taskEntry {
	ent := &taskEntry{Entry: golug_base.New(name)}
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent

}
func New(name string) Entry { return newEntry(name) }
