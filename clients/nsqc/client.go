package nsqc

import (
	"io"
	"sync"

	"github.com/pubgo/xerror"
	"github.com/segmentio/nsq-go"

	"github.com/pubgo/lava/pkg/merge"
)

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	stop []func()
}

func (t *wrapper) Close() error {
	return xerror.Try(func() {
		for i := range t.stop {
			t.stop[i]()
		}
	})
}

type Client struct {
	mu  sync.Mutex
	cfg Cfg
	v   *wrapper
}

func (t *Client) Kind() string { return Name }
func (t *Client) Close() error { return t.v.Close() }

func (t *Client) Consumer(topic string, channel string) (c *nsq.Consumer, err error) {
	defer xerror.RespErr(&err)

	t.mu.Lock()
	defer t.mu.Unlock()

	var cfg nsq.ConsumerConfig
	merge.Struct(&cfg, t.cfg)
	cfg.Topic = topic
	cfg.Channel = channel

	consumer := xerror.PanicErr(nsq.StartConsumer(cfg)).(*nsq.Consumer)
	for msg := range consumer.Messages() {
		if msg.Complete() {
			continue
		}
	}

	t.v.stop = append(t.v.stop, consumer.Stop)
	return consumer, nil
}

func (t *Client) Producer(topic string) (p *nsq.Producer, err error) {
	defer xerror.RespErr(&err)

	t.mu.Lock()
	defer t.mu.Unlock()

	var cfg nsq.ProducerConfig
	merge.Struct(&cfg, t.cfg)
	cfg.Topic = topic

	producer := xerror.PanicErr(nsq.StartProducer(cfg)).(*nsq.Producer)

	t.v.stop = append(t.v.stop, producer.Stop)
	return producer, nil
}
