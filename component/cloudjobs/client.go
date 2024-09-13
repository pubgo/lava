package cloudjobs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/component/natsclient"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/internal/ctxutil"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Params struct {
	Nc  *natsclient.Client
	Cfg *Config
	Lc  lifecycle.Lifecycle
}

func New(p Params) *Client {
	js := assert.Must1(jetstream.New(p.Nc.Conn))
	return &Client{
		p:         p,
		js:        js,
		prefix:    DefaultPrefix,
		handlers:  make(map[string]map[string]JobHandler[proto.Message]),
		streams:   make(map[string]jetstream.Stream),
		consumers: make(map[string]map[string]jetstream.Consumer),
		jobs:      make(map[string]map[string]map[string]*jobHandler),
	}
}

type Client struct {
	p  Params
	js jetstream.JetStream

	// stream manager
	streams map[string]jetstream.Stream

	// jobs: stream->consumer->Consumer
	consumers map[string]map[string]jetstream.Consumer

	// handlers: job name -> subject -> job handler
	handlers map[string]map[string]JobHandler[proto.Message]

	// jobs: stream->consumer->subject->jobHandler
	jobs map[string]map[string]map[string]*jobHandler

	// stream, consumer, subject prefix, default: DefaultPrefix
	prefix string
}

func (c *Client) initStream() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	for streamName, cfg := range c.p.Cfg.Streams {
		streamName = c.streamName(streamName)

		assert.If(c.streams[streamName] != nil, "stream %s already exists", streamName)

		// add subject prefix
		streamSubjects := lo.Map(cfg.Subjects, func(item string, index int) string { return c.subjectName(item) })
		metadata := map[string]string{"creator": fmt.Sprintf("%s/%s/%s", version.Project(), version.Version(), running.InstanceID)}
		storageType := getStorageType(cfg.Storage)
		streamCfg := jetstream.StreamConfig{
			Name:     streamName,
			Subjects: streamSubjects,
			Metadata: metadata,
			Storage:  storageType,
		}

		stream, err := c.js.CreateOrUpdateStream(ctx, streamCfg)
		if err != nil && !errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
			return errors.Wrapf(err, "failed to create stream:%s", streamName)
		}
		c.streams[streamName] = stream
	}
	return nil
}

func (c *Client) initConsumer() error {
	allEventKeysSet := mapset.NewSet(lo.MapToSlice(subjects, func(key string, value proto.Message) string { return c.subjectName(key) })...)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	for jobOrConsumerName, consumers := range c.p.Cfg.Consumers {
		jobName := jobOrConsumerName
		assert.If(c.handlers[jobName] == nil, "failed to find job handler: %s, please impl RegisterCloudJob", jobName)

		consumerName := jobOrConsumerName
		for _, cfg := range consumers {
			// check subject exists
			for _, sub := range cfg.Subjects {
				name := c.subjectName(lo.FromPtr(sub.Name))
				assert.If(!allEventKeysSet.Contains(name), "subject:%s not found, please check protobuf define and service", name)
			}

			consumerName = c.consumerName(typex.DoFunc1(func() string {
				if cfg.Consumer != nil {
					return lo.FromPtr(cfg.Consumer)
				}
				return consumerName
			}))
			streamName := c.streamName(cfg.Stream)

			// consumer init
			typex.DoFunc(func() {
				if c.consumers[streamName] == nil {
					c.consumers[streamName] = make(map[string]jetstream.Consumer)
				}
				// A streaming consumer can only have one corresponding job handler
				assert.If(c.consumers[streamName][consumerName] != nil, "consumer %s already exists", consumerName)

				metadata := map[string]string{"version": fmt.Sprintf("%s/%s", version.Project(), version.Version())}
				consumerCfg := jetstream.ConsumerConfig{
					Name:     consumerName,
					Durable:  consumerName,
					Metadata: metadata,
				}

				consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, consumerCfg)
				assert.Fn(err != nil && !errors.Is(err, jetstream.ErrConsumerExists), func() error {
					return errors.Wrapf(err, "stream=%s consumer=%s", streamName, consumerName)
				})
				logger.Info().Func(func(e *zerolog.Event) {
					e.Str("stream", streamName)
					e.Str("consumer", consumerName)
					e.Msg("register consumer success")
				})

				c.consumers[streamName][consumerName] = consumer
			})

			typex.DoFunc(func() {
				if c.jobs[streamName] == nil {
					c.jobs[streamName] = make(map[string]map[string]*jobHandler)
				}

				if c.jobs[streamName][consumerName] == nil {
					c.jobs[streamName][consumerName] = map[string]*jobHandler{}
				}

				baseJobConfig := handleDefaultJobConfig(cfg.Job)
				subjectMap := lo.SliceToMap(cfg.Subjects, func(item1 *strOrJobConfig) (string, *JobConfig) {
					item := lo.ToPtr(JobConfig(lo.FromPtr(item1)))
					return c.subjectName(*item.Name), mergeJobConfig(item, baseJobConfig)
				})

				for subName, subCfg := range subjectMap {
					assert.If(c.handlers[jobName][subName] == nil, "job handler not found, job_name=%s sub_name=%s", jobName, subName)

					job := &jobHandler{
						name:    jobName,
						handler: c.handlers[jobName][subName],
						cfg:     subCfg,
					}

					logger.Info().Func(func(e *zerolog.Event) {
						e.Str("job_name", job.name)
						e.Str("job_handler", stack.CallerWithFunc(job.handler).String())
						e.Any("job_config", subCfg)
						e.Any("stream_name", streamName)
						e.Any("consumer_name", consumerName)
						e.Any("job_subject", subName)
						e.Msg("register cloud job handler executor")
					})
					c.jobs[streamName][consumerName][subName] = job
				}
			})
		}
	}
	return nil
}

func (c *Client) doHandler(meta *jetstream.MsgMetadata, msg jetstream.Msg, job *jobHandler, cfg *JobConfig) (gErr error) {
	var timeout = lo.FromPtr(cfg.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ctx = log.UpdateEventCtx(ctx, log.Map{
		"sub_subject":          msg.Subject(),
		"sub_stream":           meta.Stream,
		"sub_consumer":         meta.Consumer,
		"sub_msg_id":           msg.Headers().Get(jetstream.MsgIDHeader),
		"sub_msg_" + senderKey: msg.Headers().Get(senderKey),
	})

	msgCtx := &Context{
		Context:      ctx,
		Header:       http.Header(msg.Headers()),
		NumDelivered: meta.NumDelivered,
		NumPending:   meta.NumPending,
		Timestamp:    meta.Timestamp,
		Stream:       meta.Stream,
		Consumer:     meta.Consumer,
		Subject:      msg.Subject(),
		Config:       cfg,
	}

	var now = time.Now()
	var args any
	defer func() {
		if gErr == nil {
			return
		}

		logger.Err(gErr).Func(func(e *zerolog.Event) {
			e.Any("context", msgCtx)
			e.Any("args", args)
			e.Str("timeout", timeout.String())
			e.Str("start_time", now.String())
			e.Str("job_cost", time.Since(now).String())
			e.Msg("failed to do cloud job handler")
		})
	}()

	var pb anypb.Any
	if err := proto.Unmarshal(msg.Data(), &pb); err != nil {
		return errors.WrapTag(err,
			errors.T("msg", "failed to unmarshal stream msg data to any proto"),
			errors.T("args", string(msg.Data())),
		)
	}
	args = &pb

	dst, err := anypb.UnmarshalNew(&pb, proto.UnmarshalOptions{})
	if err != nil {
		return errors.WrapTag(err,
			errors.T("msg", "failed to unmarshal any proto to proto msg"),
			errors.T("args", pb.String()),
		)
	}
	args = &dst

	return errors.WrapFn(job.handler(msgCtx, dst), func() errors.Tags {
		return errors.Tags{
			errors.T("msg", "failed to do cloud job handler"),
			errors.T("args", dst),
			errors.T("any_pb", pb.String()),
		}
	})
}

func (c *Client) doConsume() error {
	for streamName, consumers := range c.consumers {
		for consumerName, consumer := range consumers {
			assert.If(c.jobs[streamName] == nil, "stream not found, stream=%s", streamName)
			assert.If(c.jobs[streamName][consumerName] == nil, "consumer not found, consumer=%s", consumerName)

			jobSubjects := c.jobs[streamName][consumerName]

			logger.Info().Func(func(e *zerolog.Event) {
				e.Str("stream", streamName)
				e.Str("consumer", consumerName)
				e.Any("subjects", lo.MapKeys(jobSubjects, func(_ *jobHandler, key string) string { return key }))
				e.Msg("cloud job do consumer")
			})

			var doConsumeHandler = func(msg jetstream.Msg) {
				var now = time.Now()
				var addMsgInfo = func(e *zerolog.Event) {
					e.Str("stream", streamName)
					e.Str("consumer", consumerName)
					e.Any("header", msg.Headers())
					e.Any("msg_id", msg.Headers().Get(jetstream.MsgIDHeader))
					e.Str("subject", msg.Subject())
					e.Str("msg_received_time", now.String())
					e.Str("job_cost", time.Since(now).String())
				}

				logger.Debug().Func(func(e *zerolog.Event) {
					addMsgInfo(e)
					e.Msg("received cloud job event")
				})

				handler := jobSubjects[msg.Subject()]
				if handler == nil {
					logger.Error().Func(addMsgInfo).Msg("failed to find subject job handler")
					return
				}

				meta, err := msg.Metadata()
				if err != nil {
					// no ack, retry always, unless it can recognize special error information
					logger.Err(err).Func(addMsgInfo).Msg("failed to parse nats stream msg metadata")
					return
				}

				var cfg = handler.cfg
				var checkErrAndLog = func(err error, msg string) {
					if err == nil {
						return
					}

					logger.Err(err).
						Str("fn_caller", stack.Caller(1).String()).
						Func(addMsgInfo).
						Any("metadata", meta).
						Any("config", cfg).
						Any("msg_received_time", now.String()).
						Str("job_cost", time.Since(now).String()).
						Msg(msg)
				}

				err = try.Try(func() error { return c.doHandler(meta, msg, handler, cfg) })
				if err == nil {
					checkErrAndLog(msg.Ack(), "failed to do msg ack with handler ok")
					return
				}

				// reject job msg
				if isRejectErr(err) {
					checkErrAndLog(msg.TermWithReason("reject by caller"), "failed to do msg ack with reject err")
					return
				}

				var backoff = lo.FromPtr(cfg.RetryBackoff)
				var maxRetries = lo.FromPtr(cfg.MaxRetry)

				// Proactively retry and did not reach the maximum retry count
				if meta.NumDelivered < uint64(maxRetries) {
					logger.Warn().
						Err(err).
						Func(addMsgInfo).
						Any("metadata", meta).
						Msg("retry nats stream cloud job event")
					checkErrAndLog(msg.NakWithDelay(backoff), "failed to retry msg with delay nak")
					return
				}

				checkErrAndLog(err, "failed to do handler cloud job")
				checkErrAndLog(msg.Ack(), "failed to do msg ack with handler error")
			}

			con := result.Of(consumer.Consume(doConsumeHandler))
			if con.IsErr() {
				return errors.WrapCaller(con.Err())
			}
			c.p.Lc.BeforeStop(func() { con.Unwrap().Stop() })
		}
	}
	return nil
}

func (c *Client) Start() error {
	if err := c.initStream(); err != nil {
		return errors.WrapCaller(err)
	}

	if err := c.initConsumer(); err != nil {
		return errors.WrapCaller(err)
	}

	return errors.WrapCaller(c.doConsume())
}

func (c *Client) Publish(ctx context.Context, key string, args proto.Message, opts ...jetstream.PublishOpt) error {
	return c.publish(ctx, key, args, opts...)
}

func (c *Client) publish(ctx context.Context, key string, args proto.Message, opts ...jetstream.PublishOpt) (gErr error) {
	var timeout = ctxutil.GetTimeout(ctx)
	var now = time.Now()
	var msgId = xid.New().String()
	defer func() {
		var msgFn = func(e *zerolog.Event) {
			e.Str("pub_topic", key)
			e.Str("pub_start", now.String())
			e.Any("pub_args", args)
			e.Str("pub_cost", time.Since(now).String())
			e.Str("pub_msg_id", msgId)
			if timeout != nil {
				e.Str("timeout", timeout.String())
			}
		}
		if gErr == nil {
			logger.Info(ctx).Func(msgFn).Msg("succeed to publish cloud job event to stream")
		} else {
			logger.Err(gErr, ctx).Func(msgFn).Msg("failed to publish cloud job event to stream")
		}
	}()

	pb, err := anypb.New(args)
	if err != nil {
		return errors.Wrap(err, "failed to marshal args to any proto")
	}

	// TODO get info from ctx

	data, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrap(err, "failed to marshal any proto to bytes")
	}

	// subject|topic name
	key = c.subjectName(key)

	msg := &nats.Msg{
		Subject: key,
		Data:    data,
		Header: nats.Header{
			senderKey: []string{fmt.Sprintf("%s/%s", running.Project, running.Version)},
		},
	}

	opts = append(opts, jetstream.WithMsgID(msgId))
	_, err = c.js.PublishMsg(ctx, msg, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to publish msg to stream, topic=%s msg_id=%s", key, msgId)
	}

	return nil
}

func (c *Client) streamName(name string) string {
	prefix := fmt.Sprintf("%s:", c.prefix)
	if strings.HasPrefix(name, prefix) {
		return name
	}

	return fmt.Sprintf("%s%s", prefix, name)
}

func (c *Client) consumerName(name string) string {
	prefix := fmt.Sprintf("%s:", c.prefix)
	if strings.HasPrefix(name, prefix) {
		return name
	}

	return fmt.Sprintf("%s%s", prefix, name)
}

func (c *Client) subjectName(name string) string {
	return handleSubjectName(name, c.prefix)
}
