package rest

import (
	"context"
	"errors"
	"net/http"
	"sync"

	fb "github.com/pubgo/lug/builder/fiber"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/base"
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/types"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
)

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	*base.Entry
	cfg        Cfg
	srv        fb.Builder
	handlers   []func()
	middleOnce sync.Once
	handler    func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error
}

// Register 注册grpc handler
func (t *restEntry) Register(srv interface{}) {
	defer xerror.RespExit()

	xerror.Assert(srv == nil, "[srv] should not be nil")

	// 检查是否实现了grpc handler
	xerror.Assert(!checkHandle(srv).IsValid(), "[srv] 没有找到对应的service实现")

	t.handlers = append(t.handlers, func() {
		xerror.Panic(dix.Invoke(srv))
		xerror.PanicF(register(t.srv.Get(), srv), "[rest] grpc handler register error")
	})
}

func (t *restEntry) Start() error {
	return try.Try(func() {
		// 启动server后等待
		fx.GoDelay(func() {
			logs.Infof("Server Listening On http://localhost:%s", getPort(runenv.Addr))

			if err := t.srv.Get().Listen(runenv.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logs.Error("Server Close Error", logutil.Err(err))
				return
			}

			logs.Infof("Server Closed OK")
		})
	})
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	logs.Info("Server Shutdown")

	if err := t.srv.Get().Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logs.Error("Server Shutdown Error", logutil.Err(err))
		return err
	}

	logs.Info("Server Shutdown Ok")

	return nil
}

func newEntry(name string) *restEntry {
	var ent = &restEntry{
		Entry: base.New(name),
		srv:   fb.New(),
	}

	ent.OnInit(func() {
		defer xerror.RespExit()

		trace(ent)

		ent.cfg.DisableStartupMessage = true
		// 解析rest_entry配置
		_ = config.Decode(Name, &ent.cfg)

		// 初始化srv
		xerror.Panic(ent.srv.Build(ent.cfg.Cfg))

		// 加载组件middleware
		// lug middleware比fiber Middleware的先加载
		ent.srv.Get().Use(ent.handlerLugMiddle(ent.Options().Middlewares))

		// 依赖注入router
		xerror.Exit(dix.Provider(ent.srv.Get()))

		// 初始化router
		for i := range ent.handlers {
			ent.handlers[i]()
		}
	})

	return ent
}

func New(name string) Entry { return newEntry(name) }
