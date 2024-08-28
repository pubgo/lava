package cloudjobs

import (
	"context"
	"net/http"
	"time"
)

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
