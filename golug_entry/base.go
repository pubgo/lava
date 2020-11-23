package golug_entry

import (
	"fmt"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ Entry = (*baseEntry)(nil)

type baseEntry struct {
	opts     Options
	handlers []func()
}

func (t *baseEntry) Decode(name string, fn interface{}) (err error) {
	return xerror.Wrap(golug_config.Decode(name, fn))
}

func (t *baseEntry) Init() error {
	t.opts.Initialized = true
	golug_config.Project = t.Options().Name

	return nil
}

func (t *baseEntry) Run() RunEntry { return t }

func (t *baseEntry) Start() error { return nil }

func (t *baseEntry) Stop() error { return nil }

func (t *baseEntry) UnWrap(fn interface{}) error { panic("implement me") }

func (t *baseEntry) Options() Options { return t.opts }

func (t *baseEntry) Flags(fn func(flags *pflag.FlagSet)) (err error) {
	defer xerror.RespErr(&err)
	fn(t.opts.Command.PersistentFlags())
	return nil
}

func (t *baseEntry) Description(description ...string) error {
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

	return nil
}

func (t *baseEntry) Version(v string) error {
	t.opts.Version = strings.TrimSpace(v)
	if t.opts.Version == "" {
		return xerror.New("[version] should not be null")
	}

	t.opts.Command.Version = v
	_, err := ver.NewVersion(v)
	return xerror.WrapF(err, "[v] version format error")
}

func (t *baseEntry) Commands(commands ...*cobra.Command) error {
	rootCmd := t.opts.Command
	for _, cmd := range commands {
		if cmd == nil {
			continue
		}

		if rootCmd.Name() == cmd.Name() {
			return xerror.Fmt("command(%s) already exists", cmd.Name())
		}

		rootCmd.AddCommand(cmd)
	}
	return nil
}

func (t *baseEntry) initFlags() {
	xerror.Panic(t.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&t.opts.Addr, "addr", t.opts.Addr, "the server address")
	}))
}

func newEntry(name string) *baseEntry {
	name = strings.TrimSpace(name)
	if name == "" {
		xerror.Panic(xerror.New("the [name] parameter should not be empty"))
	}

	rootCmd := &cobra.Command{Use: name}
	runCmd := &cobra.Command{Use: "run", Short: "run as a service"}
	rootCmd.AddCommand(runCmd)

	ent := &baseEntry{
		opts: Options{
			Name:       name,
			Addr:       ":8080",
			RunCommand: runCmd,
			Command:    rootCmd,
		},
	}

	ent.initFlags()

	return ent
}

func New(name string) *baseEntry {
	return newEntry(name)
}
