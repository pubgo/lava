package cloudjobs

import (
	"fmt"
	"reflect"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/lava/pkg/proto/cloudjobpb"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

var subjects = make(map[string]proto.Message)

func RegisterSubject(subject string, subType proto.Message) any {
	assert.If(subject == "", "subject is empty")
	assert.If(subType == nil, "subType is nil")
	assert.If(subjects[subject] != nil, "subject %s already registered", subject)
	logger.Info().Func(func(e *zerolog.Event) {
		e.Str("subject", subject)
		e.Str("type", string(subType.ProtoReflect().Descriptor().FullName()))
		e.Msg("register subject")
	})

	subjects[subject] = subType
	return nil
}

func RegisterJobHandler[T proto.Message](jobCli *Client, jobName string, topic string, handler JobHandler[T], opts ...*cloudjobpb.RegisterJobOptions) {
	assert.Fn(reflect.TypeOf(subjects[topic]) != reflect.TypeOf(lo.Empty[T]()), func() error {
		return fmt.Errorf("type not match, topic-type=%s handler-input-type=%s", reflect.TypeOf(subjects[topic]).String(), reflect.TypeOf(lo.Empty[T]()).String())
	})

	jobCli.registerJobHandler(jobName, topic, func(ctx *Context, args proto.Message) error { return handler(ctx, args.(T)) }, opts...)
}

func (c *Client) registerJobHandler(jobName string, topic string, handler JobHandler[proto.Message], opts ...*cloudjobpb.RegisterJobOptions) {
	assert.If(handler == nil, "job handler is nil")
	assert.If(subjects[topic] == nil, "topic:%s not found", topic)

	if c.handlers[jobName] == nil {
		c.handlers[jobName] = map[string]JobHandler[proto.Message]{}
	}

	topic = c.subjectName(topic)
	c.handlers[jobName][topic] = handler

	logger.Info().Func(func(e *zerolog.Event) {
		e.Str("job_name", jobName)
		e.Str("topic", topic)
		e.Str("job_handler", stack.CallerWithFunc(handler).String())
		e.Msg("register cloud job handler")
	})
}
