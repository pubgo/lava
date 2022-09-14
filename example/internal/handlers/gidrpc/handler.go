package gidrpc

import (
	"context"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
)

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	return id.srv.Generate(ctx, req)
}

func (id *Id) Types(ctx context.Context, req *gidpb.TypesRequest) (*gidpb.TypesResponse, error) {
	return id.srv.Types(ctx, req)
}
