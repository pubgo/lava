package rest

import (
	"context"
	"github.com/pubgo/dix"
	"net/http"

	fb "github.com/pubgo/lug/builder/fiber"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/types"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	*base.Entry
	cfg      Cfg
	srv      fb.Builder
	handlers []func()
	cancel   context.CancelFunc
}

func (t *restEntry) handlerMiddle(fbCtx *fiber.Ctx) error {
	fbCtx.SetUserContext(fbCtx.Context())
	var wrapper = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		if err := fbCtx.Next(); err != nil {
			return xerror.Wrap(err)
		}
		return xerror.Wrap(rsp(&httpResponse{ctx: fbCtx}))
	}

	var middlewares = t.Options().Middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapper = middlewares[i](wrapper)
	}

	// create a client.Request
	request := &httpRequest{req: fbCtx}
	return wrapper(fbCtx.UserContext(), request, func(response types.Response) error { return nil })
}

func (t *restEntry) Register(handler Service, handlers ...Handler) {
	defer xerror.RespExit()

	xerror.Assert(handler == nil, "[handler] should not be nil")

	t.BeforeStart(func() { xerror.Exit(dix.Invoke(handler)) })

	var handles []interface{}
	for i := range handlers {
		handles = append(handles, handlers[i])
	}

	t.handlers = append(t.handlers, func() {
		var srv fiber.Router = t.srv.Get()
		if len(handles) > 0 {
			srv = srv.Use(handles...)
		}

		if checkHandle(handler).IsValid() {
			xerror.PanicF(register(srv, handler), "[rest] register error")
		}
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

func (t *restEntry) Start() error {
	return try.Try(func() {
		// 启动server后等待
		fx.GoDelay(func() {
			logs.Infof("Server [rest] Listening on http://localhost%s", runenv.Addr)

			if err := t.srv.Get().Listen(runenv.Addr); err != nil && err != http.ErrServerClosed {
				logs.Error("Server [rest] Close Error", zap.Any("err", err))
				return
			}

			logs.Infof("Server [rest] Closed OK")
		})
	})
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	logs.Info("Server [rest] Shutdown")
	if err := t.srv.Get().Shutdown(); err != nil && err != http.ErrServerClosed {
		logs.Error("Rpc [rest] Shutdown Error", zap.Any("err", err))
		return err
	}
	logs.Info("Server [rest] Shutdown Ok")

	return nil
}

func newEntry(name string) *restEntry {
	var ent = &restEntry{
		Entry: base.New(name),
		cfg:   Cfg{},
		srv:   fb.New(),
	}

	ent.OnInit(func() {
		defer xerror.RespExit()

		ent.cfg.DisableStartupMessage = true
		_ = config.Decode(Name, &ent.cfg)

		// 初始化srv
		xerror.Panic(ent.srv.Build(ent.cfg.Cfg))

		// 加载组件middleware
		ent.srv.Get().Use(ent.handlerMiddle)

		// 初始化routes
		for i := range ent.handlers {
			ent.handlers[i]()
		}

		ent.trace()
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }
