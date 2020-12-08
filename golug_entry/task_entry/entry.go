package task_entry

import (
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_broker/nsq_broker"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/base_entry"
	"github.com/pubgo/xerror"
)

var _ golug_entry.TaskEntry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler golug_broker.Handler
	opts    golug_broker.SubOptions
	optList []golug_broker.SubOption
}

type taskEntry struct {
	golug_entry.Entry
	cfg      Cfg
	broker   golug_broker.Broker
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) error {
	var _opts golug_broker.SubOptions
	for i := range opts {
		opts[i](&_opts)
	}
	_opts.Topic = topic

	t.handlers = append(t.handlers, entryTaskHandler{handler: handler, opts: _opts, optList: opts})
	return nil
}

func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

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

func (t *taskEntry) Stop() error { return nil }

func (t *taskEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *taskEntry) Run() golug_entry.RunEntry { return t }

func (t *taskEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_entry.UnWrap(t, fn)) }

func (t *taskEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())

	xerror.Panic(t.Decode(Name, &t.cfg))

	for _, c := range t.cfg.Consumers {
		driver := c.Driver
		if driver == "nsq" {
			t.broker = nsq_broker.NewBroker(c.Name)
			break
		} else {
			return xerror.Fmt("%s not found", c.Driver)
		}
	}

	return nil
}

func newEntry(name string) *taskEntry {
	ent := &taskEntry{Entry: base_entry.New(name)}
	ent.trace()
	return ent
}

func New(name string) *taskEntry {
	return newEntry(name)
}