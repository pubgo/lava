package task_entry

import (
	"context"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/base_entry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
)

var _ golug_entry.TaskEntry = (*taskEntry)(nil)

type entryTaskHandler struct {
	handler golug_entry.TaskHandler
	opts    golug_entry.TaskCallOptions
}

type taskEntry struct {
	golug_entry.Entry
	cfg      Cfg
	handlers []entryTaskHandler
}

func (t *taskEntry) Register(topic string, handler golug_entry.TaskHandler, opts ...golug_entry.TaskCallOption) error {
	var _opts golug_entry.TaskCallOptions
	for i := range opts {
		opts[i](&_opts)
	}
	_opts.Topic = topic

	t.handlers = append(t.handlers, entryTaskHandler{handler: handler, opts: _opts})
	return nil
}

func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	for i := range t.handlers {
		handler := t.handlers[i]
		consumer := t.cfg.c
		if handler.opts.Consumer != nil {
			consumer = handler.opts.Consumer
		}

		cancel := xprocess.Go(func(ctx context.Context) error {
			return xerror.Wrap(consumer.Subscribe(ctx, handler.opts.Topic, handler.handler))
		})
		xerror.Panic(dix_run.WithAfterStop(func(ctx *dix_run.AfterStopCtx) { xerror.Panic(cancel()) }))
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
