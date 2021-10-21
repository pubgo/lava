package nsqc

import (
	"github.com/pubgo/lava/resource"
	"sync"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/segmentio/nsq-go"
)

var _ resource.Resource = (*Client)(nil)

type Client struct {
	mu   sync.Mutex
	cfg  Cfg
	stop []func()
}

func (t *Client) Close() error {
	return xerror.Try(func() {
		for i := range t.stop {
			t.stop[i]()
		}
	})
}

func (t *Client) UpdateResObj(val interface{}) {}
func (t *Client) Kind() string                 { return Name }

func (t *Client) Consumer(topic string, channel string) (c *nsq.Consumer, err error) {
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

func (t *Client) Producer(topic string) (p *nsq.Producer, err error) {
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
