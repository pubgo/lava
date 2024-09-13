package cloudjobs

import (
	"context"
	"net/http"
	"time"

	"github.com/pubgo/lava/pkg/proto/cloudjobpb"
	"github.com/rs/xid"
	"google.golang.org/protobuf/proto"
)

var pushEventCtxKey = xid.New().String()

type Context struct {
	context.Context

	// Header jetstream.Headers().
	Header http.Header

	// NumDelivered jetstream.MsgMetadata{}.NumDelivered
	NumDelivered uint64

	// NumPending jetstream.MsgMetadata{}.NumPending
	NumPending uint64

	// Timestamp jetstream.MsgMetadata{}.Timestamp
	Timestamp time.Time

	// Stream jetstream.MsgMetadata{}.Stream
	Stream string

	// Consumer jetstream.MsgMetadata{}.Consumer
	Consumer string

	// Subject|Topic name jetstream.Msg().Subject()
	Subject string

	// Config job config from config file or default
	Config *JobConfig
}

func withOptions(ctx context.Context, opts *cloudjobpb.PushEventOptions) context.Context {
	if opts == nil {
		return ctx
	}

	oldOpts, ok := ctx.Value(pushEventCtxKey).(*cloudjobpb.PushEventOptions)
	if !ok {
		oldOpts = opts
	} else {
		proto.Merge(oldOpts, opts)
	}

	return context.WithValue(ctx, pushEventCtxKey, oldOpts)
}

func getOptions(ctx context.Context, opts ...*cloudjobpb.PushEventOptions) *cloudjobpb.PushEventOptions {
	var evtOpt = new(cloudjobpb.PushEventOptions)
	opt, ok := ctx.Value(pushEventCtxKey).(*cloudjobpb.PushEventOptions)
	if ok {
		evtOpt = opt
	}

	for _, o := range opts {
		proto.Merge(evtOpt, o)
	}

	return evtOpt
}
