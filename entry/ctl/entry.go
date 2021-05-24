package ctl

import (
	"github.com/pubgo/lug/app"
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

func (t *ctlEntry) Register(f func(), opts ...Opt) { t.handler = f }
func (t *ctlEntry) Start() (err error)             { return xutil.Try(t.handler) }

func (t *ctlEntry) Init() (err error) {
	defer xerror.RespErr(&err)
	xerror.Panic(t.Entry.Init())
	_ = config.Decode(Name, &t.cfg)
	app.IsBlock = false
	return
}

func newEntry(name string) *ctlEntry { return &ctlEntry{Entry: base.New(name)} }
func New(name string) Entry          { return newEntry(name) }
