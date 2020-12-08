package nsq_broker

import (
	"sync"

	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_plugin/plugins/golug_nsq"
	"github.com/pubgo/xerror"
	"github.com/segmentio/nsq-go"
)

var _ golug_broker.Broker = (*nsqBroker)(nil)

func NewBroker(name string) golug_broker.Broker {
	return &nsqBroker{name: name}
}

type nsqBroker struct {
	name string
	mu   sync.Mutex
	sap  sync.Map
}

func (t *nsqBroker) Publish(topic string, msg *golug_broker.Message, opts ...golug_broker.PubOption) (err error) {
	defer xerror.RespErr(&err)

	val, ok := t.sap.Load(topic)
	if ok {
		return xerror.WrapF(val.(*nsq.Producer).PublishTo(topic, msg.Body), "topic:%s", topic)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	val, ok = t.sap.Load(topic)
	if ok {
		xerror.Panic(val.(*nsq.Producer).PublishTo(topic, msg.Body))
		return nil
	}

	n, _err := golug_nsq.GetClient(t.name)
	xerror.Panic(_err)

	p := xerror.PanicErr(n.Producer(topic)).(*nsq.Producer)
	t.sap.Store(topic, p)

	xerror.Panic(p.PublishTo(topic, msg.Body))

	return nil
}

func (t *nsqBroker) Subscribe(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) (err error) {
	defer xerror.RespErr(&err)

	var _opts golug_broker.SubOptions
	for _, opt := range opts {
		opt(&_opts)
	}

	n, _err := golug_nsq.GetClient(t.name)
	xerror.Panic(_err)

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
