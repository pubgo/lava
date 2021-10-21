package restEntry

import (
	"context"
	"errors"
	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/plugins/logger"
	"net/http"
	"sync"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/base"
	fb "github.com/pubgo/lava/pkg/builder/fiber"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

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
		// lava middleware比fiber Middleware的先加载
		ent.srv.Get().Use(ent.handlerMiddle(ent.Options().Middlewares))

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
func (t *restEntry) Register(srv interface{}, opts ...Opt) {
	defer xerror.RespExit()

	xerror.Assert(srv == nil, "[srv] should not be nil")

	// 检查是否实现了handler
	xerror.Assert(!checkHandle(srv).IsValid(), "[srv] 没有找到对应的service实现")

	t.handlers = append(t.handlers, func() {
		xerror.Panic(dix.Inject(srv))

		// 如果handler实现了InitHandler接口
		if init, ok := srv.(entry.InitHandler); ok {
			init.Init()
		}

		xerror.PanicF(register(t.srv.Get(), srv), "[rest] grpc handler register error")
	})
}

func (t *restEntry) Start() error {
	return try.Try(func() {
		// 启动server后等待
		fx.GoDelay(func() {

			logz.Named(Name).Infof("Server Listening On http://localhost:%s", getPort(runenv.Addr))
			if err := t.srv.Get().Listen(runenv.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logz.Named(Name).Error("Server Close Error", logger.WithErr(err))
				return
			}

			logz.Named(Name).Info("Server Closed OK")
		})
	})
}

func (t *restEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	logz.Named(Name).Info("Server Shutdown")

	if err := t.srv.Get().Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logz.Named(Name).Error("Server Shutdown Error", logger.WithErr(err))
		return err
	}

	logz.Named(Name).Info("Server Shutdown Ok")

	return nil
}
