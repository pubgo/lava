package natsclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
	"github.com/panjf2000/ants/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/rs/xid"

	"github.com/aginetwork7/portal-server/internal/component/gopool"
	"github.com/aginetwork7/portal-server/internal/component/redis"
	pkgerror "github.com/aginetwork7/portal-server/pkg/error"
)

type Param struct {
	Cfg    *Config
	Logger log.Logger
	Lc     lifecycle.Lifecycle
	Redis  *redis.Client
}

type Client struct {
	Param

	*nats.Conn
	js nats.JetStreamContext

	subs     map[string]map[string]Subscriber
	groupMap map[string]string

	handlerPool *ants.Pool
}

func New(p Param) *Client {
	c := &Client{Param: p}
	c.Logger = c.Logger.WithName("nats")

	nc := assert.Must1(nats.Connect(c.Cfg.Url, func(o *nats.Options) error {
		o.AllowReconnect = true
		o.Name = fmt.Sprintf("%s/%s/%s", running.Hostname, running.Project, running.InstanceID)
		return nil
	}))
	log.Info().Bool("status", nc.IsConnected()).Msg("nats connection ...")

	c.Lc.BeforeStop(func() {
		nc.Close()
	})

	c.Conn = nc
	c.js = assert.Must1(c.Conn.JetStream())
	c.subs = make(map[string]map[string]Subscriber)
	c.groupMap = make(map[string]string)

	c.handlerPool = assert.Must1(ants.NewPool(
		c.Cfg.Pool.WorkerSize,
		ants.WithLogger(&gopool.Logger{Logger: c.Logger}),
		ants.WithMaxBlockingTasks(c.Cfg.Pool.QueueSize),
	))
	c.Lc.BeforeStop(func() {
		c.handlerPool.Release()
	})

	return c
}

func (c *Client) Sub(topic string, msgHandler nats.MsgHandler) {
	log.Info().Str("topic", topic).Msg("start nats subscribe")
	srvSub := assert.Must1(c.Subscribe(topic, msgHandler))
	c.Lc.BeforeStop(func() {
		err := srvSub.Unsubscribe()
		if err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			log.Err(err).Str("topic", topic).Msg("failed to unsubscribe")
		}
	})
}

func (c *Client) AddSub(sub Subscriber) {
	topic := sub.GetTopic()
	srv := sub.GetSrv()
	if _, found := c.subs[topic]; !found {
		c.subs[topic] = make(map[string]Subscriber)
	}
	if _, found := c.subs[topic][srv]; found {
		c.Logger.Fatal().Msgf("service %s subscribe topic %s more than once", srv, topic)
	}
	c.subs[topic][srv] = sub
	c.Logger.Info().Msgf("add subscriber for topic %s service %s", topic, srv)
}

func (c *Client) GetSub(topic, group string) Subscriber {
	key := topic
	subs := c.subs[key]
	if len(subs) == 0 {
		return nil
	}
	return subs[group]
}

func (c *Client) GroupSub(topic, group string, msgHandler nats.MsgHandler) {
	log.Info().Str("topic", topic).Str("group", group).Msg("start nats queue subscribe")
	srvSub := assert.Must1(c.QueueSubscribe(topic, group, msgHandler))
	c.Lc.BeforeStop(func() {
		err := srvSub.Unsubscribe()
		if err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			log.Err(err).Str("topic", topic).Str("queue", group).Msg("failed to unsubscribe")
		}
	})
}

type StreamGroupSub struct {
	Stream      string `yaml:"stream"`
	Topic       string `yaml:"topic"`
	Group       string `yaml:"group"`
	GroupPrefix string `yaml:"group_prefix"`
	BatchSize   int    `yaml:"batch_size"`
}

func (c *Client) StreamGroupSub(ctx context.Context, sub *StreamGroupSub) {
	topic := sub.Topic
	realGroup := sub.Group
	group := sub.Group
	if len(sub.GroupPrefix) > 0 {
		group = fmt.Sprintf("%s_%s", sub.GroupPrefix, group)
	}
	c.groupMap[group] = realGroup
	c.Logger.Info().Str("topic", topic).Str("group", group).Msg("start nats queue subscribe")
	cid := xid.New().String()
	assert.Must1(c.js.AddConsumer(sub.Stream, &nats.ConsumerConfig{
		Durable:        group,
		DeliverSubject: cid,
		DeliverGroup:   group,
		AckPolicy:      nats.AckExplicitPolicy,
	}))
	srvSub := assert.Must1(c.js.QueueSubscribe(
		topic, group, c.msgHandler, nats.ManualAck(), nats.Context(ctx),
		nats.AckExplicit(), nats.DeliverSubject(cid),
	))
	c.Lc.BeforeStop(func() {
		err := srvSub.Unsubscribe()
		if err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			c.Logger.Err(err).Str("topic", topic).Str("group", group).Msg("failed to unsubscribe")
		}
	})
}

func (c *Client) msgHandler(nm *nats.Msg) {
	topic := nm.Subject
	data := nm.Data
	group := nm.Sub.Queue
	realGroup := c.groupMap[group]
	logger := c.Logger.WithFields(log.Map{
		"topic":      topic,
		"data":       data,
		"group":      group,
		"real_group": realGroup,
	})
	logger.Debug().Msg("received msg")
	var needAck = false
	defer func(needAck *bool) {
		if *needAck {
			if err := nm.Ack(); err != nil {
				logger.Err(err).Msg("ack msg err")
			} else {
				logger.Debug().Msg("ack msg")
			}
		}
	}(&needAck)

	sub := c.GetSub(topic, realGroup)
	if sub == nil {
		logger.Warn().Msg("no handler")
		needAck = true
		return
	} else {
		var m msg
		err := json.Unmarshal(data, &m)
		if err != nil {
			logger.Err(err).Msg("parse data err")
			needAck = true
			return
		}

		if c.Cfg.EnableIgnore {
			exist, err := c.Redis.Exists(context.TODO(), fmt.Sprintf("nats:nm:%s", m.ID)).Result()
			if err != nil {
				logger.Err(err).Msg("check msg handled flag err")
			} else if exist > 0 {
				logger.Warn().Msg("msg is ignored")
				needAck = true
				return
			}
		}

		{
			if m.StartedAt.After(time.Now()) {
				delay := m.StartedAt.Sub(time.Now())
				if err = nm.NakWithDelay(delay); err != nil {
					logger.Err(err).Msg("nak err")
					return
				}
				logger.Debug().Msgf("delay after %f seconds", delay.Seconds())
				return
			}
		}

		err = c.handlerPool.Submit(func() {
			var needAck = false
			var err error
			var delay time.Duration
			defer func(needAck *bool) {
				if *needAck {
					if err := nm.Ack(); err != nil {
						logger.Err(err).Msg("ack msg err")
					} else {
						logger.Debug().Msg("ack msg")
					}
				}
			}(&needAck)
			needAck, delay, err = sub.Handle(topic, m.Data)
			if err != nil {
				pkgerror.LogErr(logger.Error(), err).Msg("handle msg err")
			} else if needAck {
				logger.Debug().Msg("success to handle msg")
				if c.Cfg.EnableIgnore {
					if err = c.Redis.Set(
						context.TODO(), fmt.Sprintf("nats:nm:%s", m.ID), "1", time.Hour,
					).Err(); err != nil {
						logger.Err(err).Msg("set msg handled flag err")
					}
				}
			} else if delay > 0 {
				err = nm.NakWithDelay(delay)
				if err != nil {
					logger.Err(err).Msg("nak err")
					return
				}
				logger.Debug().Msgf("delay after %f seconds to handle msg", delay.Seconds())
			}
		})
		if err != nil {
			logger.Err(err).Msg("submit to handle err")
		}
	}
}

type Stream struct {
	stream *nats.StreamInfo

	srvName string
}

type StreamCfg struct {
	Name   string
	Topics []string
}

func (c *Client) NewStream(sc *StreamCfg) *Stream {
	cfg := &nats.StreamConfig{
		Name:     sc.Name,
		Subjects: sc.Topics,
	}
	s, err := c.js.AddStream(cfg)
	if err != nil {
		if !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
			assert.Must(err)
		} else {
			assert.Must1(c.js.UpdateStream(cfg))
		}
	}
	return &Stream{
		stream:  s,
		srvName: sc.Name,
	}
}

type basePublisher struct {
	c     *Client
	topic string
}

func (p *basePublisher) GetTopic() string {
	return p.topic
}

type delayedPublisher struct {
	basePublisher
}

func (c *Client) NewPub(sc *StreamCfg, topic string) Publisher {
	c.NewStream(sc)
	p := &basePublisher{
		c:     c,
		topic: topic,
	}
	return p
}

func (c *Client) NewDelayedPub(sc *StreamCfg, topic string) DelayedPublisher {
	c.NewStream(sc)
	p := &delayedPublisher{
		basePublisher{
			c:     c,
			topic: topic,
		},
	}
	return p
}

type msg struct {
	ID        string          `json:"id"`
	StartedAt time.Time       `json:"started_at"`
	Data      json.RawMessage `json:"data"`
}

func (c *Client) newMsg(data any) (r result.Result[*msg]) {
	d, err := json.Marshal(data)
	if err != nil {
		return r.WithErr(err)
	}
	m := &msg{
		ID:   xid.New().String(),
		Data: d,
	}
	return r.WithVal(m)
}

type msgOpt func(m *msg)

func (p *basePublisher) pub(ctx context.Context, data any, opts ...msgOpt) error {
	var dataInternal *msg
	{
		t := p.c.newMsg(data)
		if t.IsErr() {
			return t.Err()
		}
		dataInternal = t.Unwrap()
	}
	for _, opt := range opts {
		opt(dataInternal)
	}
	v, err := json.Marshal(dataInternal)
	if err != nil {
		return err
	}
	_, err = p.c.js.PublishMsg(&nats.Msg{
		Subject: p.topic,
		Data:    v,
	}, nats.ContextOpt{Context: ctx})
	if p.c.Logger.Debug().Enabled() {
		p.c.Logger.Debug(ctx).Str("topic", p.topic).Str("data", string(v)).Msg("publish msg")
	}
	return err
}

func (p *basePublisher) Pub(ctx context.Context, data any) error {
	return p.pub(ctx, data)
}

func (p *delayedPublisher) PubDelayed(ctx context.Context, data any, startedAt time.Time) error {
	return p.pub(ctx, data, func(m *msg) {
		m.StartedAt = startedAt
	})
}

type baseSubscriber struct {
	c       *Client
	topic   string
	srv     string
	stream  *Stream
	handler Handler
}

func (s *baseSubscriber) GetTopic() string {
	return s.topic
}

func (s *baseSubscriber) GetSrv() string {
	return s.srv
}

func (s *baseSubscriber) Handle(topic string, m []byte) (bool, time.Duration, error) {
	return s.handler(topic, m)
}

func (c *Client) NewSub(sc *StreamCfg, topic, srv string, handler Handler) Subscriber {
	if handler == nil {
		c.Logger.Fatal().Str("stream", sc.Name).Str("topic", topic).Msg("handler is nil")
	}
	s := c.NewStream(sc)
	sub := &baseSubscriber{
		c:       c,
		topic:   topic,
		srv:     srv,
		stream:  s,
		handler: handler,
	}
	c.AddSub(sub)
	return sub
}

func (c *Client) NewDelayedSub(sc *StreamCfg, topic, srv string, handler Handler) Subscriber {
	return c.NewSub(sc, topic, srv, handler)
}
