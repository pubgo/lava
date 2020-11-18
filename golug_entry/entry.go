package golug_entry

import (
	"context"
	"fmt"
	"github.com/pubgo/golug/golug_abc"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ golug_abc.Entry = (*httpEntry)(nil)

type httpEntry struct {
	app      *fiber.App
	opts     golug_abc.Options
	handlers []func()
}

func (t *httpEntry) Group(prefix string, fn func(r fiber.Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app.Group(prefix)) })
}

func (t *httpEntry) Options() golug_abc.Options {
	return t.opts
}

func (t *httpEntry) Use(handler ...fiber.Handler) {
	for i := range handler {
		if handler[i] == nil {
			continue
		}

		i := i
		t.handlers = append(t.handlers, func() { t.app.Use(handler[i]) })
	}
}

func (t *httpEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	t.opts.Initialized = true
	golug_config.Project = t.Options().Name

	// 初始化app
	t.app = fiber.New(t.opts.RestCfg)
	return nil
}

func (t *httpEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化routes
	for i := range t.handlers {
		t.handlers[i]()
	}

	cancel := xprocess.Go(func(ctx context.Context) (err error) {
		defer xerror.RespErr(&err)

		addr := t.Options().RestAddr
		log.Infof("Server [http] Listening on http://%s", addr)
		xerror.Panic(t.app.Listen(addr))
		log.Infof("Server [http] Closed OK")

		return nil
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { xerror.Panic(cancel()) }))

	return nil
}

func (t *httpEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	if err := t.app.Shutdown(); err != nil && err != http.ErrServerClosed {
		fmt.Println(xerror.Parse(err).Println())
	}
	return nil
}

func (t *httpEntry) initCfg() {
	xerror.Panic(golug_config.Decode("server", &t.opts.RestCfg))
}

func (t *httpEntry) initFlags() {
	xerror.Panic(t.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&t.opts.RestAddr, "http_addr", t.opts.RestAddr, "the http server address")
		flags.BoolVar(&t.opts.RestCfg.DisableStartupMessage, "disable_startup_message", t.opts.RestCfg.DisableStartupMessage, "print out the http server art and listening address")
	}))
}

func (t *httpEntry) Flags(fn func(flags *pflag.FlagSet)) (err error) {
	defer xerror.RespErr(&err)
	fn(t.opts.Command.PersistentFlags())
	return nil
}

func (t *httpEntry) Description(description ...string) error {
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

func (t *httpEntry) Version(v string) error {
	t.opts.Version = strings.TrimSpace(v)
	if t.opts.Version == "" {
		return xerror.New("[version] should not be null")
	}

	t.opts.Command.Version = v
	_, err := ver.NewVersion(v)
	return xerror.WrapF(err, "[v] version format error")
}

func (t *httpEntry) Commands(commands ...*cobra.Command) error {
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

func newEntry(name string) *httpEntry {
	name = strings.TrimSpace(name)
	if name == "" {
		xerror.Panic(xerror.New("the [name] parameter should not be empty"))
	}

	rootCmd := &cobra.Command{Use: name}
	runCmd := &cobra.Command{Use: "run", Short: "run as a service"}
	rootCmd.AddCommand(runCmd)

	ent := &httpEntry{
		opts: golug_abc.Options{
			RestCfg:    fiber.New().Config(),
			Name:       name,
			RestAddr:   ":8080",
			RunCommand: runCmd,
			Command:    rootCmd,
		},
	}
	ent.initFlags()
	ent.initCfg()
	ent.trace()

	return ent
}

func New(name string) *httpEntry {
	return newEntry(name)
}

//"#${pid} - ${time} ${status} - ${latency} ${method} ${path}\n"
