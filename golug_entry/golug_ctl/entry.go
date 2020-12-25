package golug_ctl

import (
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	golug_entry.Entry
	cfg     Cfg
	handler func()
}

func (t *ctlEntry) Start() (err error)                { return xerror.Try(t.handler) }
func (t *ctlEntry) Stop() error                       { return nil }
func (t *ctlEntry) Register(f func(), opts ...Option) { t.handler = f }
func (t *ctlEntry) Options() golug_entry.Options      { return t.Entry.Run().Options() }
func (t *ctlEntry) Run() golug_entry.RunEntry         { return t }
func (t *ctlEntry) UnWrap(fn interface{})             { xerror.Next().Panic(golug_utils.UnWrap(t, fn)) }
func (t *ctlEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())
	golug_config.Decode(Name, &cfg)
	golug_app.IsBlock = false
	return nil
}

func newEntry(name string, cfg interface{}) *ctlEntry {
	ent := &ctlEntry{Entry: golug_base.New(name, cfg)}
	return ent
}

func New(name string, cfg interface{}) *ctlEntry {
	return newEntry(name, cfg)
}
