package golug_base

import (
	"fmt"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xerror/xerror_abc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ golug_entry.Entry = (*Entry)(nil)

type Entry struct {
	opts golug_entry.Options
}

func (t *Entry) WithBeforeStart(f func(_ *golug_entry.BeforeStart)) {
	xerror.Panic(dix_run.WithBeforeStart(f))
}

func (t *Entry) WithAfterStart(f func(_ *golug_entry.AfterStart)) {
	xerror.Panic(dix_run.WithAfterStart(f))
}

func (t *Entry) WithBeforeStop(f func(_ *golug_entry.BeforeStop)) {
	xerror.Panic(dix_run.WithBeforeStop(f))
}

func (t *Entry) WithAfterStop(f func(_ *golug_entry.AfterStop)) {
	xerror.Panic(dix_run.WithAfterStop(f))
}

func (t *Entry) Plugin(plugin golug_plugin.Plugin) {
	defer xerror.RespRaise(func(err xerror_abc.XErr) error { return xerror.Wrap(err, "Entry.Plugin") })
	golug_plugin.Register(plugin, golug_plugin.Module(t.opts.Name))
}

func (t *Entry) OnCfg(fn interface{}) {
	xerror.Assert(fn == nil, "[fn] is null")

	golug_config.On(func(cfg *golug_config.Config) { golug_config.Decode(t.opts.Name, fn) })
}

func (t *Entry) OnCfgWithName(name string, fn interface{}) {
	xerror.Assert(fn == nil || name == "", "[name,fn] is null")

	golug_config.On(func(cfg *golug_config.Config) { golug_config.Decode(name, fn) })
}

func (t *Entry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(golug_app.Project != t.Options().Name, "project name not match(%s, %s)", golug_app.Project, t.Options().Name)

	t.opts.Initialized = true
	return
}
func (t *Entry) Dix(data ...interface{})      { xerror.Panic(dix.Dix(data...)) }
func (t *Entry) Start() error                 { return nil }
func (t *Entry) Stop() error                  { return nil }
func (t *Entry) Options() golug_entry.Options { return t.opts }
func (t *Entry) Flags(fn func(flags *pflag.FlagSet)) {
	defer xerror.RespExit()
	fn(t.opts.Command.PersistentFlags())
}

func (t *Entry) Description(description ...string) {
	t.opts.Command.Short = fmt.Sprintf("This is a %s service", t.opts.Name)

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

func (t *Entry) Version(v string) {
	t.opts.Version = strings.TrimSpace(v)
	if t.opts.Version == "" {
		return
	}

	t.opts.Command.Version = v
	_, err := ver.NewVersion(v)
	xerror.Panic(err)
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

func (t *Entry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.UintVar(&t.opts.Port, "port", t.opts.Port, "the server port")
	})
}

func handleCmdName(name string) string {
	if strings.Contains(name, "-") {
		names := strings.Split(name, "-")
		name = names[len(names)-1]
	}

	if strings.Contains(name, ".") {
		names := strings.Split(name, ".")
		name = names[len(names)-1]
	}

	return name
}

func newEntry(name string) *Entry {
	name = strings.TrimSpace(name)
	xerror.Assert(name == "", "the [name] parameter should not be empty")
	xerror.Assert(strings.Contains(name, " "), "[name] should not contain blank")

	ent := &Entry{
		opts: golug_entry.Options{
			Name:    name,
			Port:    8080,
			Command: &cobra.Command{Use: handleCmdName(name)},
		},
	}

	ent.initFlags()

	return ent
}

func New(name string) *Entry {
	return newEntry(name)
}
