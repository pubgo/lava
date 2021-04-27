package grpc_gw

import (
	"context"
	"net/http"
	"time"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type ServeMux = gw.ServeMux

type Cfg struct {
	Timeout time.Duration `json:"timeout"`
}

func (t Cfg) Build(opts ...gw.ServeMuxOption) *gw.ServeMux {
	if t.Timeout != 0 {
		gw.DefaultContextTimeout = t.Timeout
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

	tOpts = append(tOpts, opts...)
	return gw.NewServeMux(tOpts...)
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Timeout: time.Second * 2,
	}
}