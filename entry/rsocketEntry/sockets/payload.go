package sockets

import "github.com/rsocket/rsocket-go/payload"

type ErrPayload struct {
	payload.Payload
	Err chan error
}

func NewErrPayload(data payload.Payload) *ErrPayload {
	return &ErrPayload{Payload: data, Err: make(chan error)}
}
