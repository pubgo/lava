package golug_rest

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_base"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/golug/golug_xgen"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
)

var httpMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodConnect: {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

var _ Entry = (*restEntry)(nil)

type restEntry struct {
	golug_entry.Entry
	app      *fiber.App
	handlers []func()
}

func (t *restEntry) Register(handler interface{}, opts ...golug_grpc.Option) {
	defer xerror.RespExit()

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v, data := range golug_xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
			continue
		}

		if !hd.Implements(v1.In(1)) {
			continue
		}

		var handlers = data

		vh := reflect.ValueOf(handler)
		for _, h := range handlers {
			if h.ServerStreams || h.ClientStream {
				continue
			}

			if _, ok := httpMethods[h.Method]; !ok {
				continue
			}

			h := h
			t.handlers = append(t.handlers, func() {
				mth := vh.MethodByName(h.Name)
				mthInType := mth.Type().In(1)

				t.app.Add(h.Method, h.Path, func(view *fiber.Ctx) error {
					mthIn := reflect.New(mthInType.Elem())
					ret := reflect.ValueOf(view.BodyParser).Call([]reflect.Value{mthIn})
					if !ret[0].IsNil() {
						return xerror.Wrap(ret[0].Interface().(error))
					}

					ret = mth.Call([]reflect.Value{reflect.ValueOf(view.Context()), mthIn})
					if !ret[1].IsNil() {
						return xerror.Wrap(ret[1].Interface().(error))
					}

					return xerror.Wrap(view.JSON(ret[0].Interface()))
				})
			})
		}

		return
	}
}

func (t *restEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *restEntry) Run() golug_entry.RunEntry { return t }

func (t *restEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_util.UnWrap(t, fn)) }

func (t *restEntry) Router(prefix string, fn func(r fiber.Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app.Group(prefix)) })
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

	t.Decode(Name, &cfg)

	return nil
}

func (t *restEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化app
	t.app = fiber.New(cfg)

	// 初始化routes
	for i := range t.handlers {
		t.handlers[i]()
	}

	cancel := xprocess.GoDelay(time.Second, func(ctx context.Context) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("grpcEntry.Start handle error", xlog.Any("err", err))
		})

		addr := t.Options().Addr
		xlog.Infof("Server [http] Listening on http://%s", addr)
		if err := t.app.Listen(addr); err != nil && err != http.ErrServerClosed {
			xlog.Error(xerror.Parse(err).Stack(true))
			return
		}

		xlog.Infof("Server [http] Closed OK")
		return
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { cancel() }))

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

func (t *restEntry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.BoolVar(&cfg.DisableStartupMessage, "disable_startup_message", cfg.DisableStartupMessage, "print out the http server art and listening address")
	})
}

func newEntry(name string) *restEntry {
	ent := &restEntry{Entry: golug_base.New(name)}
	ent.initFlags()
	ent.trace()
	return ent
}

func New(name string) *restEntry {
	return newEntry(name)
}
