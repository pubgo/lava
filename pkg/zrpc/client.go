package zrpc

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

// Client sends zrpc unary requests through NATS.
type Client struct {
	nc          *nats.Conn
	middlewares []lava.Middleware
}

// NewClient creates a zrpc client bound to a NATS connection.
func NewClient(nc *nats.Conn, middlewares ...lava.Middleware) *Client {
	return &Client{nc: nc, middlewares: middlewares}
}

// Conn returns the underlying NATS connection.
func (c *Client) Conn() *nats.Conn {
	return c.nc
}

// CallUnary sends a unary protobuf request and decodes the protobuf response.
func (c *Client) CallUnary(ctx context.Context, subject string, timeout time.Duration, req, resp proto.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := ctx.Deadline(); !ok && timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	header := requestHeaderFromContext(ctx)
	if len(header.ContentType()) == 0 {
		header.SetContentType(DefaultContentType)
	}
	header.SetMethod(MethodNATS)
	header.SetRequestURI(subject)
	header.Set(httputil.HeaderXRequestOperation, subject)

	reqID := firstNotEmpty(
		lavacontexts.GetReqID(ctx),
		string(header.Peek(httputil.HeaderXRequestID)),
		newRequestID(),
	)
	header.Set(httputil.HeaderXRequestID, reqID)

	if deadline, ok := ctx.Deadline(); ok {
		header.Set(HeaderTimeout, time.Until(deadline).String())
	}

	ctx = lavacontexts.CreateCtxWithReqID(ctx, reqID)
	ctx = lavacontexts.CreateReqHeader(ctx, header)

	wrappedReq := &request{
		client:      true,
		subject:     subject,
		service:     serviceFromSubject(subject),
		contentType: string(header.ContentType()),
		header:      header,
		payload:     req,
	}

	handler := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		payload, err := proto.Marshal(req.Payload().(proto.Message))
		if err != nil {
			return nil, err
		}

		msg := nats.NewMsg(subject)
		msg.Data = payload
		msg.Header = requestHeaderToNATS(req.Header())

		rspMsg, err := c.nc.RequestMsgWithContext(ctx, msg)
		if err != nil {
			return nil, err
		}

		if err = ErrorFromMessage(rspMsg); err != nil {
			return nil, err
		}

		if err = proto.Unmarshal(rspMsg.Data, resp); err != nil {
			return nil, err
		}

		wrappedRsp := &response{header: responseHeaderFromNATS(rspMsg.Header), payload: resp}
		copyResponseHeader(contextResponseHeader(ctx), wrappedRsp.header)
		return wrappedRsp, nil
	}

	_, err := lava.Chain(c.middlewares...).Middleware(handler)(ctx, wrappedReq)
	return err
}
