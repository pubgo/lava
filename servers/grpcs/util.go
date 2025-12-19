package grpcs

import (
	"context"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"google.golang.org/grpc/metadata"
)

// serviceFromMethod returns the service
// /service.Foo/Bar => service.Foo
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	return strings.Split(strings.Trim(m, "/"), "/")[0]
}

func getIncomingMetadata(ctx context.Context) metadata.MD {
	inprocCtx := inprocgrpc.ClientContext(ctx)
	if inprocCtx != nil {
		ctx = inprocCtx
	}

	reqMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		reqMetadata = make(metadata.MD)
	}
	return reqMetadata
}
