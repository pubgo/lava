package grpcc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/pubgo/lava/v2/clients/grpcc/grpccconfig"
	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lavacontexts"
)

func TestServiceFromMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		want   string
	}{
		{"", ""},
		{"no-leading-slash", "no-leading-slash"},
		{"/pkg.Service/Method", "pkg"},
		{"/a.b.c.Service/Method", "a.b.c"},
		{"/short", "/short"},
	}
	for _, tt := range tests {
		if got := serviceFromMethod(tt.method); got != tt.want {
			t.Errorf("serviceFromMethod(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestUnaryInterceptorPreservesRequestID(t *testing.T) {
	t.Parallel()

	const wantID = "req-from-ctx"
	var gotID string
	capture := lava.MiddlewareWrap{
		Name: "capture",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (lava.Response, error) {
				gotID = string(req.Header().Peek(httputil.HeaderXRequestID))
				return next(ctx, req)
			}
		},
	}

	ic := unaryInterceptor([]lava.Middleware{capture})
	ctx := lavacontexts.CreateCtxWithReqID(context.Background(), wantID)
	err := ic(
		ctx,
		"/demo.DemoService/Call",
		struct{}{},
		struct{}{},
		nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if gotID != wantID {
		t.Fatalf("request id = %q, want %q", gotID, wantID)
	}
}

func TestUnaryInterceptorContentTypeFromMetadata(t *testing.T) {
	t.Parallel()

	var gotCT string
	capture := lava.MiddlewareWrap{
		Name: "capture",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (lava.Response, error) {
				gotCT = req.ContentType()
				return next(ctx, req)
			}
		},
	}

	ic := unaryInterceptor([]lava.Middleware{capture})
	md := metadata.Pairs("x-content-type", "application/json")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	err := ic(
		ctx,
		"/demo.DemoService/Call",
		struct{}{},
		struct{}{},
		nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if gotCT != "application/json" {
		t.Fatalf("content type = %q, want application/json", gotCT)
	}
}

func TestUnaryInterceptorDefaultContentType(t *testing.T) {
	t.Parallel()

	var gotCT string
	capture := lava.MiddlewareWrap{
		Name: "capture",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (lava.Response, error) {
				gotCT = req.ContentType()
				return next(ctx, req)
			}
		},
	}

	ic := unaryInterceptor([]lava.Middleware{capture})
	err := ic(
		context.Background(),
		"/demo.DemoService/Call",
		struct{}{},
		struct{}{},
		nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if gotCT != grpccconfig.DefaultContentType {
		t.Fatalf("content type = %q, want default %q", gotCT, grpccconfig.DefaultContentType)
	}
}

func TestHead2MdRoundTrip(t *testing.T) {
	t.Parallel()

	header := httputil.NewRequestHeader()
	header.Set("X-Custom", "value")
	md := metadata.MD{}
	head2md(header, md)

	out := httputil.NewRequestHeader()
	md2Head(md, out)

	if string(out.Peek("X-Custom")) != "value" {
		t.Fatalf("got %q", out.Peek("X-Custom"))
	}
}
