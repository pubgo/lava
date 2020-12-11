package http_entry

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/base_entry"
	"github.com/pubgo/golug/golug_entry/grpc_entry"
	"github.com/pubgo/golug/golug_xgen"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
)

const defaultContentType = "application/json"

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

var _ HttpEntry = (*httpEntry)(nil)

type httpEntry struct {
	golug_entry.Entry
	cfg      Cfg
	app      *fiber.App
	handlers []func()
}

func (t *httpEntry) Register(handler interface{}, opts ...grpc_entry.GrpcOption) {
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

func (t *httpEntry) Options() golug_entry.Options { return t.Entry.Run().Options() }

func (t *httpEntry) Run() golug_entry.RunEntry { return t }

func (t *httpEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_entry.UnWrap(t, fn)) }

func (t *httpEntry) Router(prefix string, fn func(r fiber.Router)) {
	t.handlers = append(t.handlers, func() { fn(t.app.Group(prefix)) })
}

func (t *httpEntry) use(handler fiber.Handler) {
	if handler == nil {
		return
	}

	t.handlers = append(t.handlers, func() { t.app.Use(handler) })
}

func (t *httpEntry) Use(handler ...fiber.Handler) {
	for i := range handler {
		t.use(handler[i])
	}
}

func (t *httpEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())

	xerror.Panic(t.Decode(Name, &t.cfg))

	return nil
}

func (t *httpEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化app
	t.app = fiber.New(t.cfg)

	// 初始化routes
	for i := range t.handlers {
		t.handlers[i]()
	}

	cancel := xprocess.Go(func(ctx context.Context) {
		defer xerror.RespErr(&err)

		addr := t.Options().Addr
		log.Infof("Server [http] Listening on http://%s", addr)
		if err := t.app.Listen(addr); err != nil && err != http.ErrServerClosed {
			log.Error(xerror.Parse(err).Stack(true))
			return
		}

		log.Infof("Server [http] Closed OK")
		return
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { cancel() }))

	return nil
}

func (t *httpEntry) Stop() (err error) {
	defer xerror.RespErr(&err)

	if err := t.app.Shutdown(); err != nil && err != http.ErrServerClosed {
		log.Error(xerror.Parse(err).Stack(true))
		return nil
	}

	return nil
}

func (t *httpEntry) initFlags() {
	t.Flags(func(flags *pflag.FlagSet) {
		flags.BoolVar(&t.cfg.DisableStartupMessage, "disable_startup_message", t.cfg.DisableStartupMessage, "print out the http server art and listening address")
	})
}

func newEntry(name string) *httpEntry {
	ent := &httpEntry{Entry: base_entry.New(name), cfg: fiber.New().Config()}
	ent.initFlags()
	ent.trace()
	return ent
}

func New(name string) *httpEntry {
	return newEntry(name)
}
