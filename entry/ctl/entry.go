package ctl

import (
	"context"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	*base.Entry
	cfg      Cfg
	handlers map[string]options
}

func (t *ctlEntry) RegisterLoop(fn func(ctx context.Context), optList ...Opt) {
	var opts = register(t, fn, optList...)
	t.handlers[opts.Name] = opts
}

func (t *ctlEntry) Register(fn func(ctx context.Context), optList ...Opt) {
	var opts = register(t, fn, optList...)
	opts.once = true

	t.handlers[opts.Name] = opts
}

func (t *ctlEntry) Start(args ...string) (err error) {
	defer xerror.RespErr(&err)

	var name = ""
	if len(args) > 0 {
		name = args[0]
	}

	var opts, ok = t.handlers[name]
	xerror.Assert(!ok, "%s not found", name)

	if opts.once {
		opts.cancel = fx.Go(opts.handler)
	} else {
		opts.cancel = fx.GoLoop(opts.handler)
	}

	return nil
}

func (t *ctlEntry) Stop() error {
	for _, opt := range t.handlers {
		if opt.cancel != nil {
			opt.cancel()
		}
	}
	return nil
}

func newEntry(name string) *ctlEntry {
	var ent = &ctlEntry{Entry: base.New(name), handlers: make(map[string]options)}
	ent.OnInit(func() {
		_ = config.Decode(Name, &ent.cfg)
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }

func register(t *ctlEntry, fn func(ctx context.Context), optList ...Opt) options {
	var opts = options{handler: fn}
	for i := range optList {
		optList[i](&opts)
	}

	xerror.Assert(t.handlers[opts.Name].handler != nil, "name [%s] already exists", opts.Name)
	return opts
}
