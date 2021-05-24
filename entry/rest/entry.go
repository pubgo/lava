package rest

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	fb "github.com/pubgo/lug/service/fiber"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	*base.Entry
	app      *fiber.App
	handlers []func()
	cfg      fb.Cfg
	cancel   context.CancelFunc
}

func (t *restEntry) Router(fn func(r Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app) })
}

func (t *restEntry) use(handler Handler) {
	if handler == nil {
		return
	}

	t.handlers = append(t.handlers, func() { t.app.Use(handler) })
}

func (t *restEntry) Use(handler ...Handler) {
	for i := range handler {
		t.use(handler[i])
	}
}

func (t *restEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Init())
	var cfg = fb.GetDefaultCfg()
	cfg.DisableStartupMessage = true
	_ = config.Decode(Name, &cfg)
	return
}

func (t *restEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化app
	t.app = t.cfg.Build()

	// 初始化routes
	for i := range t.handlers {
		t.handlers[i]()
	}

	// 启动server后等待1s
	xerror.Panic(fx.GoDelay(time.Second, func() {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("restEntry.Start error", xlog.Any("err", err))
		})

		for {
			if err := t.app.Listen(t.Options().Addr); err != nil && err != http.ErrServerClosed {
				if strings.Contains(err.Error(), "address already in use") {
					continue
				}

				xlog.Error(xerror.Parse(err).Stack(true))
			}
			break
		}

		xlog.Infof("Server [http] Closed OK")
	}))
	xlog.Infof("Server [http] Listening on http://%s", t.Options().Addr)

	return nil
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	if err := t.app.Shutdown(); err != nil && err != http.ErrServerClosed {
		xlog.Error(xerror.Parse(err).Stack(true))
		return nil
	}

	return nil
}

func newEntry(name string) *restEntry {
	ent := &restEntry{Entry: base.New(name)}
	ent.trace()
	return ent
}

func New(name string) Entry { return newEntry(name) }
