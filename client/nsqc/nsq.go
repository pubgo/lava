package nsqc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/segmentio/nsq-go"

	"sync"
)

type nsqClient struct {
	mu   sync.Mutex
	cfg  Cfg
	stop []func()
}

func (t *nsqClient) Consumer(topic string, channel string) (c *nsq.Consumer, err error) {
	defer xerror.RespErr(&err)

	t.mu.Lock()
	defer t.mu.Unlock()

	var cfg nsq.ConsumerConfig
	xerror.Panic(merge.MapStruct(&cfg, t.cfg))
	cfg.Topic = topic
	cfg.Channel = channel

	consumer := xerror.PanicErr(nsq.StartConsumer(cfg)).(*nsq.Consumer)
	for msg := range consumer.Messages() {
		if msg.Complete() {
			continue
		}
	}

	t.stop = append(t.stop, consumer.Stop)
	return consumer, nil
}

func (t *nsqClient) Producer(topic string) (p *nsq.Producer, err error) {
	defer xerror.RespErr(&err)

	t.mu.Lock()
	defer t.mu.Unlock()

	var cfg nsq.ProducerConfig
	xerror.Panic(merge.CopyStruct(&cfg, t.cfg))
	cfg.Topic = topic

	producer := xerror.PanicErr(nsq.StartProducer(cfg)).(*nsq.Producer)

	t.stop = append(t.stop, producer.Stop)
	return producer, nil
}
