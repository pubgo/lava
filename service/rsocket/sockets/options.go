package sockets

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type CallOption interface {
	grpc.CallOption
	Lava()
}

var _ CallOption = (*CallOptions)(nil)

// CallOptions represents the state of in-effect grpc.CallOptions.
type CallOptions struct {
	grpc.EmptyCallOption
	// Headers is a slice of metadata pointers which should all be set when
	// response header metadata is received.
	Headers []*metadata.MD
	// Trailers is a slice of metadata pointers which should all be set when
	// response trailer metadata is received.
	Trailers []*metadata.MD
	// Peer is a slice of peer pointers which should all be set when the
	// remote peer is known.
	Peer []*peer.Peer
	// Creds are per-RPC credentials to use for a call.
	Creds credentials.PerRPCCredentials
	// MaxRecv is the maximum number of bytes to receive for a single message
	// in a call.
	MaxRecv int
	// MaxSend is the maximum number of bytes to send for a single message in
	// a call.
	MaxSend int

	ContentSubtype string

	CompressorType string

	Codec encoding.Codec
}

func (co *CallOptions) Lava() {}

// SetHeaders sets all accumulated header addresses to the given metadata. This
// satisfies grpc.Header call options.
func (co *CallOptions) SetHeaders(md metadata.MD) {
	for _, hdr := range co.Headers {
		*hdr = md
	}
}

// SetTrailers sets all accumulated trailer addresses to the given metadata.
// This satisfies grpc.Trailer call options.
func (co *CallOptions) SetTrailers(md metadata.MD) {
	for _, tlr := range co.Trailers {
		*tlr = md
	}
}

// SetPeer sets all accumulated peer addresses to the given peer. This satisfies
// grpc.Peer call options.
func (co *CallOptions) SetPeer(p *peer.Peer) {
	for _, pr := range co.Peer {
		*pr = *p
	}
}

// GetCallOptions converts the given slice of grpc.CallOptions into a
// CallOptions struct.
func GetCallOptions(opts ...grpc.CallOption) *CallOptions {
	var copts CallOptions
	for _, o := range opts {
		switch o := o.(type) {
		case grpc.HeaderCallOption:
			copts.Headers = append(copts.Headers, o.HeaderAddr)
		case grpc.TrailerCallOption:
			copts.Trailers = append(copts.Trailers, o.TrailerAddr)
		case grpc.PeerCallOption:
			copts.Peer = append(copts.Peer, o.PeerAddr)
		case grpc.PerRPCCredsCallOption:
			copts.Creds = o.Creds
		case grpc.MaxRecvMsgSizeCallOption:
			copts.MaxRecv = o.MaxRecvMsgSize
		case grpc.MaxSendMsgSizeCallOption:
			copts.MaxSend = o.MaxSendMsgSize
		case grpc.CompressorCallOption:
		case grpc.ContentSubtypeCallOption:
		case grpc.CustomCodecCallOption:
		case grpc.ForceCodecCallOption:
		//	TODO 完善
		case CallOption:

		}
	}
	return &copts
}
