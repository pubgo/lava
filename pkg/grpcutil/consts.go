package grpcutil

import "time"

const (
	// DefaultContentType is the gRPC wire content type every lava transport falls
	// back to when a request carries none.
	DefaultContentType = "application/grpc"
	DefaultTimeout     = time.Second * 2
)

// Metadata keys lava passes across the gRPC boundary. The gateway's RPC
// middleware (servers/gatewayserver) and the client interceptors
// (clients/grpcc) read and strip the same entries, so the spellings live here
// rather than in each caller.
const (
	// MdContentType is the caller's HTTP content type, so a lava Middleware can see
	// it before the codec is chosen. The chain strips it from the metadata it
	// forwards, since "content-type" is the authoritative entry.
	MdContentType = "x-content-type"
	// MdRemote carries the peer address to the handler chain.
	MdRemote = "remote"
	// MdTimeout carries a request timeout as a Go duration string.
	MdTimeout = "timeout"
	// MdURL carries the original request URL for HTTP-facing operations.
	MdURL = "url"
)
