package golug_ctl

import (
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/xerror"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	*golug_base.Entry
	cfg     Cfg
	handler func()
}

func (t *ctlEntry) Start() (err error)                { return xerror.Try(t.handler) }
func (t *ctlEntry) Register(f func(), opts ...Option) { t.handler = f }
func (t *ctlEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Init())
	golug_app.IsBlock = false
	return nil
}

func newEntry(name string) *ctlEntry {
	ent := &ctlEntry{Entry: golug_base.New(name)}
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent
}

func New(name string) Entry { return newEntry(name) }
