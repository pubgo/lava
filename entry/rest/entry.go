package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	fb "github.com/pubgo/lug/builder/fiber"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	*base.Entry
	cfg      Cfg
	srv      fb.Builder
	handlers []func()
	cancel   context.CancelFunc
}

func (t *restEntry) Register(handler interface{}, handlers ...Handler) {
	defer xerror.RespExit()

	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(!checkHandle(handler).IsValid(), "register [%#v] 没有找到匹配的interface", handler)

	var handles []interface{}
	for i := range handlers {
		handles = append(handles, handlers[i])
	}

	t.handlers = append(t.handlers, func() {
		var srv fiber.Router = t.srv.Get()
		if len(handles) > 0 {
			srv = srv.Use(handles...)
		}
		xerror.PanicF(register(srv, handler), "[rest] register error")
	})
}

func (t *restEntry) Router(fn func(r Router)) {
	t.handlers = append(t.handlers, func() {
		fn(t.srv.Get())
	})
}

func (t *restEntry) use(handler Handler) {
	if handler == nil {
		return
	}

	t.handlers = append(t.handlers, func() {
		t.srv.Get().Use(handler)
	})
}

func (t *restEntry) Use(handler ...Handler) {
	for i := range handler {
		t.use(handler[i])
	}
}

func (t *restEntry) Start(args ...string) error {
	// 启动server后等待
	return fx.GoDelay(time.Millisecond*10, func() {
		xlog.Infof("Srv [rest] Listening on http://localhost%s", runenv.Addr)

		if err := t.srv.Get().Listen(runenv.Addr); err != nil && err != http.ErrServerClosed {
			xlog.Error("Srv [rest] Close Error", xlog.Any("err", err))
			return
		}

		xlog.Infof("Srv [rest] Closed OK")
	})
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	xlog.Info("Srv [rest] Shutdown")
	if err := t.srv.Get().Shutdown(); err != nil && err != http.ErrServerClosed {
		xlog.Error("Srv [rest] Shutdown Error", xlog.Any("err", err))
		return err
	}
	xlog.Info("Srv [rest] Shutdown Ok")

	return nil
}

func newEntry(name string) *restEntry {
	var ent = &restEntry{
		Entry: base.New(name),
		cfg:   Cfg{},
		srv:   fb.New(),
	}

	ent.OnInit(func() {
		defer xerror.Raise(func(err xerror.XErr) error { return err })

		ent.cfg.DisableStartupMessage = true
		_ = config.Decode(Name, &ent.cfg)

		// 初始化srv
		xerror.Panic(ent.srv.Build(ent.cfg.Cfg))

		// 初始化routes
		for i := range ent.handlers {
			ent.handlers[i]()
		}

		ent.trace()
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }
