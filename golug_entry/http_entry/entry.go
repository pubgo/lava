package http_entry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
)

var _ golug_entry.HttpEntry = (*httpEntry)(nil)

type httpEntry struct {
	golug_entry.Entry
	cfg      Cfg
	app      *fiber.App
	handlers []func()
}

func (t *httpEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *httpEntry) Run() golug_entry.RunEntry { return t }

func (t *httpEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_entry.UnWrap(t, fn)) }

func (t *httpEntry) Group(prefix string, fn func(r fiber.Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app.Group(prefix)) })
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

	xerror.Panic(t.Entry.Run().Init())

	xerror.Panic(golug_config.Decode(Name, &t.cfg))

	// 初始化app
	t.app = fiber.New(t.cfg)
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

		addr := t.Entry.Run().Options().Addr
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

func (t *httpEntry) initFlags() {
	xerror.Panic(t.Flags(func(flags *pflag.FlagSet) {
		flags.BoolVar(&t.cfg.DisableStartupMessage, "disable_startup_message", t.cfg.DisableStartupMessage, "print out the http server art and listening address")
	}))
}

func newEntry(name string) *httpEntry {
	ent := &httpEntry{
		Entry: golug_entry.New(name),
	}
	ent.initFlags()
	ent.trace()

	return ent
}

func New(name string) *httpEntry {
	return newEntry(name)
}

//"#${pid} - ${time} ${status} - ${latency} ${method} ${path}\n"