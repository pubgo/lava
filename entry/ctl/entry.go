package ctl

import (
	"context"
	"fmt"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	name      string
	listNames bool
	*base.Entry
	cfg      Cfg
	handlers map[string]options
}

func (t *ctlEntry) Register(run Service, optList ...Opt) {
	defer xerror.RespExit()

	xerror.Assert(run == nil, "[run] should not be nil")

	t.BeforeStart(func() { xerror.Exit(dix.Invoke(run)) })

	runners := run.Run()
	if runners != nil {
		for name, v := range runners {
			var opts = register(t, v, append(optList, withName(name))...)
			opts.once = true
			t.handlers[opts.Name] = opts
		}
	}

	loops := run.RunLoop()
	if loops != nil {
		for name, v := range loops {
			var opts = register(t, v, append(optList, withName(name))...)
			t.handlers[opts.Name] = opts
		}
	}
}

func (t *ctlEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	if t.listNames {
		for k := range t.handlers {
			fmt.Println("name:", k)
		}
		runenv.Block = false
		return
	}

	var opts, ok = t.handlers[t.name]
	xerror.Assert(!ok, "%s not found", t.name)

	if opts.once {
		runenv.Block = false
		opts.cancel = fx.Go(func(ctx context.Context) { opts.handler(fx.Ctx{Context: ctx}) })
		return
	}

	opts.cancel = fx.GoLoop(func(ctx fx.Ctx) { opts.handler(ctx) })

	return
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
	ent.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&ent.name, "name", consts.Default, "ctl name")
		flags.BoolVar(&ent.listNames, "list", false, "list ctl name")
	})

	ent.OnInit(func() {
		_ = config.Decode(Name, &ent.cfg)
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }

func register(t *ctlEntry, run Handler, optList ...Opt) options {
	var opts = options{handler: run}
	for i := range optList {
		optList[i](&opts)
	}

	if opts.Name == "" {
		opts.Name = consts.Default
	}

	xerror.Assert(t.handlers[opts.Name].handler != nil, "handler [%s] already exists", opts.Name)
	return opts
}
