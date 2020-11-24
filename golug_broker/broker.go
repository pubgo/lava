package golug_broker

import (
	"context"
	"crypto/tls"
)

type Broker interface {
	Options() Options
	Publish(topic string, msg *Message, opts ...PubOption) error
	Subscribe(topic string, handler Handler, opts ...SubOption) error
	String() string
}

type Option func(*Options)
type Options struct {
	Addrs     []string
	Secure    bool
	TLSConfig *tls.Config
	Context   context.Context
}

type PubOption func(*PubOptions)
type PubOptions struct {
	Context context.Context
}

type SubOption func(*SubOptions)
type SubOptions struct {
	Ctx     context.Context
	Topic   string
	Queue   string
	AutoAck bool
	Broker  Broker
}

type Handler func(*Message) error
type Message struct {
	Header    map[string]string
	ID        []byte
	Body      []byte
	Timestamp int64
	Attempts  uint16
	Priority  uint8
	ReplyTo   string
}
