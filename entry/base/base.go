package base

import (
	"fmt"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
)

func New(name string) *Entry { return newEntry(name) }

func newEntry(name string) *Entry {
	name = strings.TrimSpace(name)
	xerror.Assert(name == "", "[name] should not be null")
	xerror.Assert(strings.Contains(name, " "), "[name] should not contain blank")

	return &Entry{opts: entry.Opts{
		Name:    name,
		Command: &cobra.Command{Use: name, Version: version.Version},
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

func (t *Entry) BeforeStart(f func()) { t.opts.BeforeStarts = append(t.opts.BeforeStarts, f) }
func (t *Entry) BeforeStop(f func())  { t.opts.BeforeStops = append(t.opts.BeforeStops, f) }
func (t *Entry) AfterStart(f func())  { t.opts.AfterStarts = append(t.opts.AfterStarts, f) }
func (t *Entry) AfterStop(f func())   { t.opts.AfterStops = append(t.opts.AfterStops, f) }
func (t *Entry) Start() error         { panic("start unimplemented") }
func (t *Entry) Stop() error          { panic("stop unimplemented") }
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

// Plugin 注册插件
func (t *Entry) Plugin(plugins ...plugin.Plugin) {
	defer xerror.RespExit()

	for _, plg := range plugins {
		xerror.Assert(plg == nil || plg.Id() == "", "[plg] should not be nil")
		xerror.Assert(t.opts.Name == "", "please init project name")
		plugin.Register(plg, plugin.Module(t.opts.Name))
	}
}

func (t *Entry) InitRT() {
	defer xerror.RespExit()

	xerror.Assert(runenv.Project != t.Options().Name, "project name not match(%s, %s)", runenv.Project, t.Options().Name)
	xerror.Assert(t.init == nil, "init is nil")

	// 执行entry的init
	t.init()
}

func (t *Entry) Flags(fn func(flags *pflag.FlagSet)) {
	defer xerror.RespExit()
	fn(t.opts.Command.PersistentFlags())
}

func (t *Entry) Description(description ...string) {
	t.opts.Command.Short = fmt.Sprintf("%s service", t.opts.Name)

	if len(description) > 0 {
		t.opts.Command.Short = description[0]
	}

	if len(description) > 1 {
		t.opts.Command.Long = description[1]
	}

	if len(description) > 2 {
		t.opts.Command.Example = description[2]
	}

	return
}

func (t *Entry) Commands(commands ...*cobra.Command) {
	rootCmd := t.opts.Command
	for _, cmd := range commands {
		if cmd == nil {
			continue
		}

		if rootCmd.Name() == cmd.Name() {
			return
		}

		rootCmd.AddCommand(cmd)
	}
}
