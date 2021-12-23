package grpc_gw

import (
	"context"
	"net/http"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type Builder struct {
	name string
	mux  *gw.ServeMux
	opts []gw.ServeMuxOption
}

func (t *Builder) Get() *gw.ServeMux { return t.mux }
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
		// header处理
		gw.WithIncomingHeaderMatcher(func(key string) (string, bool) { return key, true }),
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

	return nil
}

func New(name string) Builder {
	return Builder{name: name}
}
