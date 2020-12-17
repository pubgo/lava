package golug_base

import (
	"fmt"
	"reflect"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_version"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ golug_entry.Entry = (*baseEntry)(nil)

type baseEntry struct {
	opts golug_entry.Options
}

func (t *baseEntry) Dix(data ...interface{}) {
	xerror.Next().Panic(dix.Dix(data...))
}

func (t *baseEntry) Init() error {
	t.opts.Initialized = true
	golug_env.Project = t.Options().Name
	return xerror.Wrap(golug_config.Init())
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
		fmt.Println(golug_util.MarshalIndent(golug_config.GetCfg().AllSettings()))
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
			fmt.Println(name, golug_util.MarshalIndent(v))
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
	if !strings.Contains(name, "_") {
		return name
	}

	return strings.Join(strings.Split(name, "_")[1:], "_")
}

func newEntry(name string) *baseEntry {
	name = strings.TrimSpace(name)
	if name == "" {
		xerror.Panic(xerror.New("the [name] parameter should not be empty"))
	}

	ent := &baseEntry{
		opts: golug_entry.Options{
			Name:    name,
			Addr:    ":8080",
			Command: &cobra.Command{Use: handleCmdName(name)},
		},
	}

	if golug_env.IsDev() || golug_env.IsTest() {
		ent.Commands(ent.pluginCmd())
		ent.Commands(ent.configCmd())
		ent.Commands(ent.dixCmd())
		ent.Commands(ent.verCmd())
	}

	ent.initFlags()

	return ent
}

func New(name string) *baseEntry {
	return newEntry(name)
}
