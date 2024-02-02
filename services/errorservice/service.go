package errorservice

import (
	"context"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/proto/errcodepb"
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

func (s service) Middlewares() []lava.Middleware {
	return nil
}

func (s service) ServiceDesc() *grpc.ServiceDesc {
	return &errcodepb.ErrorService_ServiceDesc
}
