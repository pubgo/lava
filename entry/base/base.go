package base

import (
	"fmt"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/dix"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ entry.Entry = (*Entry)(nil)

type Entry struct {
	opts entry.Options
}

func (t *Entry) BeforeStart(f func())    { t.opts.BeforeStarts = append(t.opts.BeforeStarts, f) }
func (t *Entry) AfterStart(f func())     { t.opts.AfterStarts = append(t.opts.AfterStarts, f) }
func (t *Entry) BeforeStop(f func())     { t.opts.BeforeStops = append(t.opts.BeforeStops, f) }
func (t *Entry) AfterStop(f func())      { t.opts.AfterStops = append(t.opts.AfterStops, f) }
func (t *Entry) Dix(data ...interface{}) { xerror.Panic(dix.Dix(data...)) }
func (t *Entry) Start() error            { return nil }
func (t *Entry) Stop() error             { return nil }
func (t *Entry) Options() entry.Options  { return t.opts }

// Plugin 注册插件
func (t *Entry) Plugin(plg plugin.Plugin) {
	defer xerror.RespExit()
	xerror.Assert(plg == nil || plg.String() == "", "[plg] should not be nil")
	xerror.Assert(t.opts.Name == "", "please init project name first")
	plugin.Register(plg, plugin.Module(t.opts.Name))
}

// OnCfg 项目配置加载解析
func (t *Entry) OnCfg(fn interface{}) { t.OnCfgWithName(t.opts.Name, fn) }
func (t *Entry) OnCfgWithName(name string, fn interface{}) {
	xerror.Assert(fn == nil || name == "", "[name,fn] should not be null")

	config.On(func(cfg *config.Config) { _ = config.Decode(name, fn) })
}

func (t *Entry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(config.Project != t.Options().Name, "project name not match(%s, %s)", config.Project, t.Options().Name)
	t.opts.Port = config.Port
	t.opts.Initialized = true
	return
}

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
	xerror.Assert(name == "", "[name] should not be null")
	xerror.Assert(strings.Contains(name, " "), "[name] should not contain blank")

	ent := &Entry{
		opts: entry.Options{
			Name:    name,
			Port:    8080,
			Version: version.Version,
			Command: &cobra.Command{Use: handleCmdName(name)},
		},
	}

	return ent
}

func New(name string) *Entry {
	return newEntry(name)
}
