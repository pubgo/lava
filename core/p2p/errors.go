package p2p

import "errors"

var (
	ErrSignalingClosed = errors.New("p2p: signaling closed")
	ErrPeerNotFound    = errors.New("p2p: peer not found")
	ErrICEFailed       = errors.New("p2p: ice connection failed")
	ErrTimeout         = errors.New("p2p: timeout")
)
