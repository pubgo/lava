package golug_task

import (
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
)

var _ Entry = (*taskEntry)(nil)

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

func (t *taskEntry) Stop() error { return nil }

func (t *taskEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *taskEntry) Run() golug_entry.RunEntry { return t }

func (t *taskEntry) UnWrap(fn interface{}) { xerror.Next().Panic(golug_util.UnWrap(t, fn)) }

func (t *taskEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())
	golug_config.Decode(Name, &t.cfg)
	if i := t.Options().Init; i != nil {
		i()
	}

	return nil
}

func newEntry(name string) *taskEntry {
	ent := &taskEntry{Entry: golug_base.New(name)}
	ent.trace()
	return ent
}

func New(name string) *taskEntry {
	return newEntry(name)
}
