package golug_broker_nsq

import (
	"sync"

	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/plugins/golug_nsq"
	"github.com/pubgo/xerror"
)

var _ golug_broker.Broker = (*nsqBroker)(nil)

type nsqBroker struct {
	name string
	mu   sync.Mutex
	data sync.Map
}

func (t *nsqBroker) Publish(topic string, msg *golug_broker.Message, opts ...golug_broker.PubOption) (err error) {
	defer xerror.RespErr(&err)

	val, ok := t.data.Load(topic)
	if ok {
		return xerror.WrapF(val.(*nsq.Producer).PublishTo(topic, msg.Body), "topic:%s", topic)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	val, ok = t.data.Load(topic)
	if ok {
		xerror.Panic(val.(*nsq.Producer).PublishTo(topic, msg.Body))
		return nil
	}

	n := golug_nsq.GetClient(t.name)

	p := xerror.PanicErr(n.Producer(topic)).(*nsq.Producer)
	t.data.Store(topic, p)

	xerror.Panic(p.PublishTo(topic, msg.Body))

	return nil
}

func (t *nsqBroker) Subscribe(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) (err error) {
	defer xerror.RespErr(&err)

	var _opts golug_broker.SubOptions
	for _, opt := range opts {
		opt(&_opts)
	}

	n := golug_nsq.GetClient(t.name)

	c := xerror.PanicErr(n.Consumer(topic, _opts.Queue)).(*nsq.Consumer)
	for msg := range c.Messages() {
		xerror.Panic(handler(&golug_broker.Message{
			ID:        []byte(msg.ID.String()),
			Body:      msg.Body,
			Timestamp: msg.Timestamp.Unix(),
			Attempts:  msg.Attempts,
		}))
		msg.Finish()
	}

	return nil
}

func (t *nsqBroker) String() string {
	return "nsq"
}
