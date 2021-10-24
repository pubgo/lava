package grpc_gw

import (
	"context"
	"github.com/rs/cors"
	"net/http"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type Builder struct {
	name string
	srv  *http.Server
	mux  *gw.ServeMux
	opts []gw.ServeMuxOption
}

func (t *Builder) Get() *http.Server { return t.srv }
func (t *Builder) Register(conn *grpc.ClientConn, handler interface{}) (err error) {
	return Register(context.Background(), t.mux, conn, handler)
}

func (t *Builder) Build(cfg *Cfg, opts ...gw.ServeMuxOption) error {
	t.opts = opts

	if cfg.Timeout != 0 {
		gw.DefaultContextTimeout = cfg.Timeout
	}

	var tOpts []gw.ServeMuxOption
	tOpts = append(tOpts,
		gw.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			return metadata.MD(r.URL.Query())
		}),

		gw.WithMarshalerOption(gw.MIMEWildcard, &gw.HTTPBodyMarshaler{
			Marshaler: &gw.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	tOpts = append(tOpts, t.opts...)

	t.mux = gw.NewServeMux(tOpts...)
	t.srv = &http.Server{Handler: cors.Default().Handler(t.mux)}

	return nil
}

func New(name string) Builder {
	return Builder{name: name}
}
