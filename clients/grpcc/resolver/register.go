package resolver

import "google.golang.org/grpc/resolver"

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}
