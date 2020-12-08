package ctl_entry

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/base_entry"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
)

var _ golug_entry.CtlEntry = (*ctlEntry)(nil)

type ctlEntry struct {
	golug_entry.Entry
	cfg     Cfg
	handler func()
}

func (t *ctlEntry) Start() (err error) {
	defer xerror.RespErr(&err)
	golug_env.IsBlock = false
	t.handler()
	return nil
}

func (t *ctlEntry) Stop() error { return nil }

func (t *ctlEntry) Register(f func(), opts ...golug_entry.CtlOption) { t.handler = f }

func (t *ctlEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *ctlEntry) Run() golug_entry.RunEntry { return t }

func (t *ctlEntry) UnWrap(fn interface{}) error { return nil }

func (t *ctlEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())

	xerror.Panic(t.Decode(Name, &t.cfg))

	return nil
}

func newEntry(name string) *ctlEntry {
	ent := &ctlEntry{Entry: base_entry.New(name)}
	ent.trace()
	return ent
}

func New(name string) *ctlEntry {
	return newEntry(name)
}
