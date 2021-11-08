package cliEntry

import (
	"context"
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/entry/base"
	"github.com/pubgo/lava/runenv"
)

var _ Entry = (*ctlEntry)(nil)

type ctlEntry struct {
	*base.Entry
	name      string
	listNames bool
	cfg       Cfg
	srv       Service
	handlers  map[string]options
}

func (t *ctlEntry) Register(srv Service, optList ...Opt) {
	defer xerror.RespExit()

	xerror.Assert(srv == nil, "[srv] should not be nil")
	t.srv = srv

	runners := srv.Run()
	if runners != nil {
		for name, v := range runners {
			var opts = register(t, v, append(optList, withName(name))...)
			opts.once = true
			t.handlers[opts.Name] = opts
		}
	}

	loops := srv.RunLoop()
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

	// 依赖注入
	xerror.Exit(dix.Inject(t.srv))

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
	var ent = &ctlEntry{
		Entry:    base.New(name),
		handlers: make(map[string]options),
	}

	ent.Flags(&cli.StringFlag{
		Name:        "name",
		Value:       consts.Default,
		Destination: &ent.name,
		Usage:       "cliEntry name",
	})

	ent.Flags(&cli.BoolFlag{
		Name:        "list",
		Value:       false,
		Destination: &ent.listNames,
		Usage:       "list cliEntry name",
	})

	ent.OnInit(func() {
		_ = config.Decode(Name, &ent.cfg)
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }
