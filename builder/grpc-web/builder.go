package grpcWeb

import (
	"fmt"
	"net/http"

	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

type Builder struct {
	routes    map[string]string
	resources map[string]struct{}
	server    *http.Server
	srv       *grpc.Server
}

func (t *Builder) Get() *http.Server { return t.server }
func (t *Builder) initRoutes() {
	for _, v := range xgen.List() {
		if v == nil {
			continue
		}

		for i := range v {
			t.routes[v[i].Path] = v[i].Service + "/" + v[i].Name
		}
	}
}
func (t *Builder) initResources() {
	for name, info := range t.srv.GetServiceInfo() {
		for _, mth := range info.Methods {
			t.resources[fmt.Sprintf("/%s/%s", name, mth.Name)] = struct{}{}
		}
	}
}
func (t *Builder) initMiddleware() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if p, ok := t.routes[req.URL.Path]; ok {
			req.URL.Path = p
		}

		if _, ok := t.resources[req.URL.Path]; !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s not found", req.URL.Path)
			return
		}

		t.srv.ServeHTTP(newGrpcWebResponse(w), req2GrpcRequest(req))
	})
}

func (t *Builder) Build(cfg Cfg, srv *grpc.Server) error {
	xerror.Assert(srv == nil, "srv is nil")

	t.srv = srv

	t.initResources()
	t.initRoutes()
	t.initMiddleware()

	return nil
}

func New() Builder {
	return Builder{
		server:    &http.Server{},
		routes:    map[string]string{},
		resources: map[string]struct{}{},
	}
}
