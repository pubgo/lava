package base

import (
	"fmt"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

func New(name string) *Entry { return newEntry(name) }

func newEntry(name string) *Entry {
	name = strings.TrimSpace(name)
	xerror.Assert(name == "", "[name] should not be null")
	xerror.Assert(strings.Contains(name, " "), "[name] should not contain blank")

	return &Entry{opts: entry.Opts{
		Name:    name,
		Command: &cli.Command{Name: name},
	}}
}

var _ entry.Entry = (*Entry)(nil)
var _ entry.Runtime = (*Entry)(nil)

type Entry struct {
	kind        string
	init        func()
	opts        entry.Opts
	middlewares []types.Middleware
}

func (t *Entry) RegisterHandler(h entry.Handler) { t.opts.Handlers = append(t.opts.Handlers, h) }
func (t *Entry) BeforeStart(f func())            { t.opts.BeforeStarts = append(t.opts.BeforeStarts, f) }
func (t *Entry) BeforeStop(f func())             { t.opts.BeforeStops = append(t.opts.BeforeStops, f) }
func (t *Entry) AfterStart(f func())             { t.opts.AfterStarts = append(t.opts.AfterStarts, f) }
func (t *Entry) AfterStop(f func())              { t.opts.AfterStops = append(t.opts.AfterStops, f) }
func (t *Entry) Start() error                    { panic("start unimplemented") }
func (t *Entry) Stop() error                     { panic("stop unimplemented") }
func (t *Entry) Options() entry.Opts {
	var opts = t.opts
	opts.Middlewares = append(t.middlewares[:len(t.middlewares):len(t.middlewares)], t.opts.Middlewares...)
	return opts
}

func (t *Entry) OnInit(init func()) { t.init = init }

func (t *Entry) Middleware(middleware types.Middleware) {
	if middleware == nil {
		return
	}

	t.opts.Middlewares = append(t.opts.Middlewares, middleware)
}

func (t *Entry) MiddlewareInter(middleware types.Middleware) {
	if middleware == nil {
		return
	}

	t.middlewares = append(t.middlewares, middleware)
}

func (t *Entry) InitRT() {
	defer xerror.RespExit()

	xerror.Assert(runtime.Project != t.Options().Name, "project name not match(%s, %s)", runtime.Project, t.Options().Name)
	xerror.Assert(t.init == nil, "init is nil")

	// 执行entry的init
	t.init()
}

func (t *Entry) Flags(flag cli.Flag) {
	if flag == nil {
		return
	}

	t.opts.Command.Flags = append(t.opts.Command.Flags, flag)
}

func (t *Entry) Description(description ...string) {
	t.opts.Command.Usage = fmt.Sprintf("%s service", t.opts.Name)

	if len(description) > 0 {
		t.opts.Command.Usage = description[0]
	}

	if len(description) > 1 {
		t.opts.Command.UsageText = description[1]
	}

	if len(description) > 2 {
		t.opts.Command.Description = description[2]
	}

	return
}

func (t *Entry) Commands(cmd *cli.Command) {
	if cmd == nil {
		return
	}

	t.opts.Command.Subcommands = append(t.opts.Command.Subcommands, cmd)
}
