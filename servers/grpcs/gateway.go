package grpcs

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

import _ "github.com/twitchtv/twirp"

func init() {
	mux := runtime.NewServeMux()
	_ = mux
	runtime.WithErrorHandler()
	runtime.WithIncomingHeaderMatcher()

}

// https://twitchtv.github.io/twirp/docs/hooks.html
