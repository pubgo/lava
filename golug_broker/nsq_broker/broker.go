package nsq_broker

import (
	"sync"

	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/plugins/golug_nsq"
	"github.com/pubgo/xerror"
	"github.com/segmentio/nsq-go"
)

var _ golug_broker.Broker = (*nsqBroker)(nil)

func NewBroker(name string) golug_broker.Broker {
	return &nsqBroker{name: name}
}

type nsqBroker struct {
	name string
	pMap sync.Map
}

func (t *nsqBroker) Options() golug_broker.Options {
	return golug_broker.Options{}
}

func (t *nsqBroker) Publish(topic string, msg *golug_broker.Message, opts ...golug_broker.PubOption) (err error) {
	defer xerror.RespErr(&err)

	val, ok := t.pMap.Load(topic)
	if ok {
		xerror.Panic(val.(*nsq.Producer).PublishTo(topic, msg.Body))
		return nil
	}

	n, _err := golug_nsq.GetNsq(t.name)
	xerror.Panic(_err)

	p := xerror.PanicErr(n.Producer(topic)).(*nsq.Producer)
	t.pMap.Store(topic, p)

	xerror.Panic(p.PublishTo(topic, msg.Body))

	return nil
}

func (t *nsqBroker) Subscribe(topic string, handler golug_broker.Handler, opts ...golug_broker.SubOption) (err error) {
	defer xerror.RespErr(&err)

	n, _err := golug_nsq.GetNsq(t.name)
	xerror.Panic(_err)

	c := xerror.PanicErr(n.Consumer(topic, "")).(*nsq.Consumer)
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
