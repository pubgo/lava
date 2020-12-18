package golug_nsq

import (
	"sync"

	"github.com/imdario/mergo"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	nsq "github.com/segmentio/nsq-go"
)

var clientM sync.Map

func GetClient(names ...string) *nsqClient {
	var name = golug_consts.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", name))
	}

	return val.(*nsqClient)
}

type nsqClient struct {
	mu   sync.Mutex
	cfg  ClientCfg
	stop []func()
}

func initClient(name string, cfg ClientCfg) {
	clientM.Store(name, &nsqClient{cfg: cfg})
}

func (t *nsqClient) Consumer(topic string, channel string) (c *nsq.Consumer, err error) {
	defer xerror.RespErr(&err)

	t.mu.Lock()
	defer t.mu.Unlock()

	var cfg nsq.ConsumerConfig
	xerror.Panic(mergo.Map(&cfg, t.cfg))
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
	xerror.Panic(mergo.Map(&cfg, t.cfg))
	cfg.Topic = topic

	producer := xerror.PanicErr(nsq.StartProducer(cfg)).(*nsq.Producer)

	t.stop = append(t.stop, producer.Stop)
	return producer, nil
}
