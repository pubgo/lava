package ginEntry

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/base"
	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugins/syncx"
	"github.com/pubgo/lava/runenv"
)

func New(name string) Entry { return newEntry(name) }

func newEntry(name string) *ginEntry {
	var ent = &ginEntry{Entry: base.New(name), srv: gin.New()}

	ent.OnInit(func() {
		defer xerror.RespExit()

		trace(ent)

		_ = config.Decode(Name, &ent.cfg)

		// 外部配置更新到gin
		merge.Struct(&ent.srv, ent.cfg)

		// 加载组件middleware
		// lava middleware比gin Middleware的先加载
		ent.srv.Use(ent.handlerMiddle(ent.Options().Middlewares))

		// 初始化router
		for _, h := range ent.Options().Handlers {
			// 依赖注入handler
			xerror.Panic(dix.Inject(h))

			xerror.PanicF(register(ent.srv, h), "[gin] grpc handler register error")

			// 初始化router
			if _h, ok := h.(Router); ok {
				_h.Router(ent.srv)
			}

			// handler初始化
			h.Init()
		}
	})

	return ent
}

var logs = logz.New(Name)
var _ Entry = (*ginEntry)(nil)

type ginEntry struct {
	*base.Entry
	cfg Cfg
	srv *gin.Engine
}

func (t *ginEntry) Register(handler entry.InitHandler) {
	defer xerror.RespExit()

	xerror.Assert(handler == nil, "[handler] should not be nil")

	// 检查是否实现了 <Router>
	xerror.Assert(!checkHandle(handler).IsValid(), "[handler] 没有找到对应的service实现")
	t.RegisterHandler(handler)
}

func (t *ginEntry) Start() error {
	return xerror.Try(func() {
		// 启动server后等待
		syncx.GoDelay(func() {
			logs.Infof("Server Listening On http://localhost:%s", netutil.MustGetPort(runenv.Addr))
			logs.Logs("Server Close", func() error {
				if err := t.srv.Run(runenv.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}
				return nil
			})
		})
	})
}

func (t *ginEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	logs.Info("Server Shutdown")

	//if err := t.srv.Get().Shutdown(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//	logs.Error("Server Shutdown Error", logger.WithErr(err))
	//	return err
	//}

	logs.Info("Server Shutdown Ok")

	return nil
}