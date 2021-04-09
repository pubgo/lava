package ctlEntry

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/x/xutil"
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
	return xutil.Try(t.handler)
}

func (t *ctlEntry) Init() (err error) {
	return xutil.Try(func() {
		xerror.Panic(t.Entry.Init())
		config.IsBlock = false
	})
}

func newEntry(name string) *ctlEntry {
	ent := &ctlEntry{Entry: base.New(name)}
	ent.OnCfgWithName(Name, &ent.cfg)
	return ent
}

func New(name string) Entry { return newEntry(name) }
