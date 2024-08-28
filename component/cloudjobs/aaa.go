package cloudjobs

import (
	"github.com/pubgo/funk/log"
	"google.golang.org/protobuf/proto"
)

var logger = log.GetLogger("cloud_jobs")

type Register interface {
	RegisterCloudJobs(jobCli *Client)
}

type JobHandler[T proto.Message] func(ctx *Context, args T) error
