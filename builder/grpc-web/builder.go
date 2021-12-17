package grpcWeb

import (
	"fmt"
	"net/http"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/xgen"
)

var logs = logz.Component(Name)

type Builder struct {
	name      string
	routes    map[string]string
	resources map[string]struct{}
	server    *http.Server
	srv       *grpc.Server
}

func (t *Builder) Get() *http.Server { return t.server }
func (t *Builder) initRoutes() {
	for _, vs := range xgen.List() {
		if vs == nil {
			continue
		}

		var v = vs.([]xgen.GrpcRestHandler)
		for i := range v {
			var name = v[i].Method + " " + v[i].Path
			xerror.Assert(t.routes[name] != "", "service [%s] route [%s] already exists", t.name, name)

			t.routes[name] = v[i].Service + "/" + v[i].Name
		}
	}
}
func (t *Builder) initResources() {
	for name, info := range t.srv.GetServiceInfo() {
		for _, mth := range info.Methods {
			t.resources[fmt.Sprintf("%s/%s", name, mth.Name)] = struct{}{}
		}
	}
}
func (t *Builder) initMiddleware() {
	t.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if p, ok := t.routes[req.URL.Path]; ok {
			req.URL.Path = p
		}

		logs.Info(req.URL.Path)
		if _, ok := t.resources[req.URL.Path]; !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "service [%s] url [%s] not found", t.name, req.URL.Path)
			return
		}

		t.srv.ServeHTTP(newGrpcWebResponse(w), req2GrpcRequest(req))
	})
}

func (t *Builder) Build(cfg *Cfg, srv *grpc.Server) error {
	xerror.Assert(srv == nil, "srv is nil")

	t.srv = srv

	t.initResources()
	t.initRoutes()
	t.initMiddleware()

	logs.Infow("build", zap.Any("routes", t.routes), zap.Any("resources", t.resources))
	return nil
}

func New(name string) Builder {
	return Builder{
		name:      name,
		server:    &http.Server{},
		routes:    make(map[string]string),
		resources: make(map[string]struct{}),
	}
}
