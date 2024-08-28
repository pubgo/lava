package eventjob

import (
	"github.com/pubgo/funk/log"
	"google.golang.org/protobuf/proto"
)

var logger = log.GetLogger("eventjob")

type Register interface {
	RegisterAsyncJob(jobCli *Client)
}

type JobHandler[T proto.Message] func(ctx *Context, args T) error
