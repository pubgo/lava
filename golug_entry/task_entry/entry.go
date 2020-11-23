package ctl_entry

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

var _ golug_entry.TaskEntry = (*taskEntry)(nil)

type taskEntry struct {
	golug_entry.Entry
	cfg      Cfg
	handlers []golug_entry.TaskHandler
}

func (t *taskEntry) Register(handler golug_entry.TaskHandler, opts ...golug_entry.TaskOption) error {
	t.handlers = append(t.handlers, handler)
	return nil
}

func (t *taskEntry) Start() (err error) {
	defer xerror.RespErr(&err)
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
	ent := &taskEntry{Entry: golug_entry.New(name)}
	ent.trace()
	return ent
}

func New(name string) *taskEntry {
	return newEntry(name)
}
