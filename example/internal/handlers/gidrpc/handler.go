package gidrpc

import (
	"context"

	"github.com/pubgo/lava/errors"

	"github.com/pubgo/lava/example/gen/proto/gidpb"
)

var ee = errors.New("hello")

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	ee.Err(nil).StatusBadRequest()
	return id.srv.Generate(ctx, req)
}

func (id *Id) Types(ctx context.Context, req *gidpb.TypesRequest) (*gidpb.TypesResponse, error) {
	return id.srv.Types(ctx, req)
}
