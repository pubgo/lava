package golug_rabbitmq

import (
	"github.com/streadway/amqp"
)

// 定义回调函数
type Handler func(amqp.Delivery) error

// 定义publish opts
type PublishOptions struct{}

type PublishOption func(*PublishOptions)

// 定义subscribe
type SubscribeOptions struct{}

type SubscribeOption func(*SubscribeOptions)
