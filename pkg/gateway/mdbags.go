package gateway

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type mdBagKey struct{}

// mdBag collects response metadata produced by Backend interceptors
// (e.g. lava middleware writing response headers) so Dispatcher can merge
// it into the headers/trailers returned to frontends.
type mdBag struct {
	header  metadata.MD
	trailer metadata.MD
}

func withMDBag(ctx context.Context) (context.Context, *mdBag) {
	if bag, ok := ctx.Value(mdBagKey{}).(*mdBag); ok && bag != nil {
		return ctx, bag
	}
	bag := &mdBag{
		header:  metadata.MD{},
		trailer: metadata.MD{},
	}
	return context.WithValue(ctx, mdBagKey{}, bag), bag
}

func mdBagFrom(ctx context.Context) *mdBag {
	bag, _ := ctx.Value(mdBagKey{}).(*mdBag)
	return bag
}

// AppendBackendHeader records a response header from a Backend interceptor.
func AppendBackendHeader(ctx context.Context, key string, values ...string) {
	if bag := mdBagFrom(ctx); bag != nil {
		bag.header.Append(key, values...)
	}
}

// AppendBackendTrailer records a response trailer from a Backend interceptor.
func AppendBackendTrailer(ctx context.Context, key string, values ...string) {
	if bag := mdBagFrom(ctx); bag != nil {
		bag.trailer.Append(key, values...)
	}
}

func mergeMD(dst metadata.MD, extra metadata.MD) metadata.MD {
	if len(extra) == 0 {
		return dst
	}
	if dst == nil {
		return metadata.Join(metadata.MD{}, extra)
	}
	return metadata.Join(dst, extra)
}
