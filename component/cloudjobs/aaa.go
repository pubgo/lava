package cloudjobs

import (
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/pkg/proto/cloudjobpb"
	"google.golang.org/protobuf/proto"
)

var logger = log.GetLogger("cloud_jobs")

type Register interface {
	RegisterCloudJobs(jobCli *Client)
}

type JobHandler[T proto.Message] func(ctx *Context, args T) error

type Options = cloudjobpb.PushEventOptions

type PushEventOpt func(opts *Options)
