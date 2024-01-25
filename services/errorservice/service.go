package errorservice

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/lava"
	errcodepb "github.com/pubgo/lava/pkg/proto/services/errcode"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func New() lava.GrpcRouter {
	return new(service)
}

type service struct {
}

func (s service) Codes(ctx context.Context, empty *emptypb.Empty) (*errcodepb.ErrCodes, error) {
	return &errcodepb.ErrCodes{
		Codes: errors.GetErrCodes(),
	}, nil
}

func (s service) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
	//TODO implement me
	panic("implement me")
}

func (s service) Middlewares() []lava.Middleware {
	//TODO implement me
	panic("implement me")
}

func (s service) ServiceDesc() *grpc.ServiceDesc {
	//TODO implement me
	panic("implement me")
}
