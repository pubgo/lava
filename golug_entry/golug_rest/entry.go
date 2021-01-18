package golug_rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
)

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	golug_entry.Entry
	app      *fiber.App
	handlers []func()
	cfg1     Cfg
	cfg      fiber.Config
	cancel   context.CancelFunc
}

func (t *restEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }
func (t *restEntry) Run() golug_entry.RunEntry    { return t }
func (t *restEntry) Router(fn func(r fiber.Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app) })
}
func (t *restEntry) use(handler fiber.Handler) {
	if handler == nil {
		return
	}

	t.handlers = append(t.handlers, func() { t.app.Use(handler) })
}

func (t *restEntry) Use(handler ...fiber.Handler) {
	for i := range handler {
		t.use(handler[i])
	}
}

func (t *restEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())
	dm := golug_config.GetCfg().GetStringMap(Name)
	delete(dm, "views")

	golug_utils.Mergo(&t.cfg, dm)

	if t.cfg1.Views.Dir != "" && t.cfg1.Views.Ext != "" {
		t.cfg.Views = html.New(t.cfg1.Views.Dir, t.cfg1.Views.Ext)
	}
	return nil
}

func (t *restEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化app
	t.app = fiber.New(t.cfg)

	// 初始化routes
	for i := range t.handlers {
		t.handlers[i]()
	}

	port := t.Options().Port

	// 启动server后等待1s
	t.cancel = xprocess.GoDelay(time.Second, func(ctx context.Context) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("restEntry.Start error", xlog.Any("err", err))
		})

		for {
			if err := t.app.Listen(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
				if strings.Contains(err.Error(), "address already in use") {
					port += 1
					continue
				}

				xlog.Error(xerror.Parse(err).Stack(true))
			}
			break
		}

		xlog.Infof("Server [http] Closed OK")
	})
	xlog.Infof("Server [http] Listening on http://%d", port)

	return nil
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	if err := t.app.Shutdown(); err != nil && err != http.ErrServerClosed {
		xlog.Error(xerror.Parse(err).Stack(true))
		return nil
	}

	t.cancel()

	return nil
}

func (t *restEntry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.BoolVar(&t.cfg1.DisableStartupMessage, "disable_startup_message", t.cfg1.DisableStartupMessage, "print out the http server art and listening address")
	})
}

func newEntry(name string) *restEntry {
	ent := &restEntry{Entry: golug_base.New(name), cfg1: GetDefaultCfg()}
	ent.initFlags()
	ent.trace()

	golug_config.On(func(cfg *golug_config.Config) { golug_config.Decode(Name, &ent.cfg1) })
	return ent
}

func New(name string) *restEntry { return newEntry(name) }
