package grpc_gw

import (
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"context"
	"net/http"
)

type Builder struct {
	name string
	srv  *http.Server
	mux  *gw.ServeMux
	opts []gw.ServeMuxOption
}

func (t *Builder) Get() *http.Server { return t.srv }
func (t *Builder) Register(conn *grpc.ClientConn) (err error) {
	return Register(context.Background(), t.mux, conn)
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
					UseProtoNames:  true,
					UseEnumNumbers: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}))

	tOpts = append(tOpts, t.opts...)

	t.mux = gw.NewServeMux(tOpts...)
	t.srv = &http.Server{Handler: t.mux}

	return nil
}

func New(name string) Builder {
	return Builder{name: name}
}
