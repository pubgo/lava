package grpc_gw

import (
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"context"
	"net/http"
)

type Builder struct {
	gw   *gw.ServeMux
	opts []gw.ServeMuxOption
}

func (t Builder) Get() *gw.ServeMux { return t.gw }
func (t Builder) Build(cfg Cfg,opts ...gw.ServeMuxOption) error {
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
	t.gw = gw.NewServeMux(tOpts...)

	return nil
}

func New() Builder {
	return Builder{}
}
