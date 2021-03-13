package broker

import (
	"context"
)

type Factory func(cfg map[string]interface{}) (Broker, error)
type Broker interface {
	Pub(topic string, msg *Message, opts *PubOpts) error
	Sub(topic string, handler Handler, opts *SubOpts) error
	String() string
}

type PubOpts struct {
	Ctx context.Context
}

type SubOpts struct {
	Ctx     context.Context
	Topic   string
	Queue   string
	AutoAck bool
	Broker  Broker
}

type Handler func(msg *Message) error
type Message struct {
	Header    map[string]string
	ID        string
	Body      []byte
	Timestamp int64
	Attempts  uint16
	Priority  uint8
	ReplyTo   string
}
