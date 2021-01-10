package golug_base

import (
	"fmt"
	"reflect"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_version"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ golug_entry.Entry = (*baseEntry)(nil)

type baseEntry struct {
	cfg  interface{}
	opts golug_entry.Options
}

func (t *baseEntry) Dix(data ...interface{}) {
	xerror.Next().Panic(dix.Dix(data...))
}

func (t *baseEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(golug_app.Project != t.Options().Name, "please set project flag")

	t.opts.Initialized = true

	if t.cfg != nil {
		golug_config.Decode(golug_app.Project, t.cfg)
	}

	// 开启trace
	if golug_app.Trace {
		xerror.Panic(dix_trace.Trigger())
	}
	return
}

func (t *baseEntry) Run() golug_entry.RunEntry { return t }

func (t *baseEntry) Start() error { return nil }

func (t *baseEntry) Stop() error { return nil }

func (t *baseEntry) UnWrap(fn interface{}) { panic("implement me") }

func (t *baseEntry) Options() golug_entry.Options { return t.opts }

func (t *baseEntry) Flags(fn func(flags *pflag.FlagSet)) {
	defer xerror.Resp(func(err xerror.XErr) { xlog.Error("baseEntry.Flags", xlog.Any("err", err)) })
	fn(t.opts.Command.PersistentFlags())
}

func (t *baseEntry) Description(description ...string) {
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

func (t *baseEntry) Version(v string) {
	t.opts.Version = strings.TrimSpace(v)
	if t.opts.Version == "" {
		return
	}

	t.opts.Command.Version = v
	_, err := ver.NewVersion(v)
	xerror.Next().Panic(err)
	return
}

func (t *baseEntry) Commands(commands ...*cobra.Command) {
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
	return
}

func (t *baseEntry) pluginCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "plugin info"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		for k, plugins := range golug_plugin.All() {
			fmt.Println("plugin namespace:", k)
			for _, p := range plugins {
				fmt.Println("plugin:", p.String(), reflect.TypeOf(p).PkgPath())
			}
		}
	}
	return cmd
}

func (t *baseEntry) configCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "cfg", Short: "config info"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println(golug_utils.MarshalIndent(golug_config.GetCfg().AllSettings()))
	}
	return cmd
}

func (t *baseEntry) dixCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "dix", Short: "dix dependency graph"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println(dix.Graph())
	}
	return cmd
}

func (t *baseEntry) verCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "ver", Short: "version info"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		for name, v := range golug_version.List() {
			fmt.Println(name, golug_utils.MarshalIndent(v))
		}
	}
	return cmd
}

func (t *baseEntry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&t.opts.Addr, "addr", t.opts.Addr, "the server address")
	})
}

func handleCmdName(name string) string {
	if !strings.Contains(name, ".") {
		return name
	}

	names := strings.Split(name, ".")
	return names[len(names)-1]
}

func newEntry(name string, cfg interface{}) *baseEntry {
	name = strings.TrimSpace(name)
	xerror.Assert(name == "", "the [name] parameter should not be empty")
	xerror.Assert(strings.Contains(name, " "), "[name] should not contain blank")
	if cfg != nil {
		xerror.Assert(reflect.TypeOf(cfg).Kind() != reflect.Ptr, "[cfg] type kind should be ptr")
	}

	golug_app.Project = name

	ent := &baseEntry{
		cfg: cfg,
		opts: golug_entry.Options{
			Name:    name,
			Addr:    ":8080",
			Command: &cobra.Command{Use: handleCmdName(name)},
		},
	}

	if golug_app.IsDev() || golug_app.IsTest() {
		ent.Commands(ent.pluginCmd())
		ent.Commands(ent.configCmd())
		ent.Commands(ent.dixCmd())
		ent.Commands(ent.verCmd())
	}

	ent.initFlags()
	ent.trace()

	return ent
}

func New(name string, cfg interface{}) *baseEntry {
	return newEntry(name, cfg)
}
