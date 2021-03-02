package golug_ctl

import (
	"github.com/pubgo/golug/entry/base"
	"github.com/pubgo/golug/golug"
	"github.com/pubgo/xerror"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	*base.Entry
	cfg     Cfg
	handler func()
}

func (t *ctlEntry) Register(f func(), opts ...Option) { t.handler = f }

func (t *ctlEntry) Start() (err error) {
	defer xerror.RespErr(&err)
	t.handler()
	return
}

func (t *ctlEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Init())
	golug.IsBlock = false
	return nil
}

func newEntry(name string) *ctlEntry {
	ent := &ctlEntry{Entry: base.New(name)}
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent
}

func New(name string) Entry { return newEntry(name) }
