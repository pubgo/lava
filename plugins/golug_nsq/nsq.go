package golug_nsq

import (
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	nsq "github.com/segmentio/nsq-go"
)

var nsqM sync.Map

func GetNsq(names ...string) (*nsqClient, error) {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}
	val, ok := nsqM.Load(name)
	if !ok {
		return nil, xerror.Fmt("%s not found", name)
	}

	return val.(*nsqClient), nil
}

type nsqClient struct {
	mu   sync.Mutex
	cfg  Cfg
	stop []func()
}

func initNsq(name string, cfg Cfg) {
	nsqM.Store(name, &nsqClient{cfg: cfg})
}

func delNsq(name string) {
	n, _ := GetNsq(name)
	if n == nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	for i := range n.stop {
		n.stop[i]()
	}
	nsqM.Delete(name)
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

type Cfg struct {
	Enabled        bool          `yaml:"enabled" json:"enabled" toml:"enabled"`
	Topic          string        `json:"topic"`
	Channel        string        `json:"channel"`
	Address        string        `json:"address"`
	Lookup         []string      `json:"lookup"`
	MaxInFlight    int           `json:"max_in_flight"`
	MaxConcurrency int           `json:"max_concurrency"`
	DialTimeout    time.Duration `json:"dial_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	DrainTimeout   time.Duration `json:"drain_timeout"`
}
