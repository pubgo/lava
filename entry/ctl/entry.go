package ctl

import (
	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/x/try"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	*base.Entry
	cfg     Cfg
	handler func()
}

func (t *ctlEntry) Register(f func(), opts ...Opt) { t.handler = f }
func (t *ctlEntry) Start() (err error)             { return try.Try(t.handler) }
func (t *ctlEntry) Stop() error                    { return nil }

func newEntry(name string) *ctlEntry {
	var ent = &ctlEntry{Entry: base.New(name)}
	ent.OnInit(func() {
		_ = config.Decode(Name, &ent.cfg)
		app.IsBlock = false
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }
