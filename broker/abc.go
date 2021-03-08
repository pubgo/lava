package broker

import (
	"context"
)

type Broker interface {
	Publish(topic string, msg *Message, opts *PubOpts) error
	Subscribe(topic string, handler Handler, opts *SubOpts) error
	Start() string
	Stop() string
	Name() string
}

type PubOpts struct {
	Context context.Context
}

type SubOpts struct {
	Ctx     context.Context
	Topic   string
	Queue   string
	AutoAck bool
	Broker  Broker
}

type Handler func(*Message) error
type Message struct {
	Header    map[string]string
	ID        string
	Body      []byte
	Timestamp int64
	Attempts  uint16
	Priority  uint8
	ReplyTo   string
}
