package fiber

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// NewServerMetadataContext creates a new context with ServerMetadata
func NewServerMetadataContext(ctx context.Context, md ServerMetadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

// ServerTransportStream implements grpc.ServerTransportStream.
// It should only be used by the generated files to support grpc.SendHeader
// outside of gRPC server use.
type ServerTransportStream struct {
	mu      sync.Mutex
	header  metadata.MD
	trailer metadata.MD
}

// Method returns the method for the stream.
func (s *ServerTransportStream) Method() string {
	return ""
}

// Header returns the header metadata of the stream.
func (s *ServerTransportStream) Header() metadata.MD {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.header.Copy()
}

// SetHeader sets the header metadata.
func (s *ServerTransportStream) SetHeader(md metadata.MD) error {
	if md.Len() == 0 {
		return nil
	}

	s.mu.Lock()
	s.header = metadata.Join(s.header, md)
	s.mu.Unlock()
	return nil
}

// SendHeader sets the header metadata.
func (s *ServerTransportStream) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

// Trailer returns the cached trailer metadata.
func (s *ServerTransportStream) Trailer() metadata.MD {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trailer.Copy()
}

// SetTrailer sets the trailer metadata.
func (s *ServerTransportStream) SetTrailer(md metadata.MD) error {
	if md.Len() == 0 {
		return nil
	}

	s.mu.Lock()
	s.trailer = metadata.Join(s.trailer, md)
	s.mu.Unlock()
	return nil
}

func timeoutDecode(s string) (time.Duration, error) {
	size := len(s)
	if size < 2 {
		return 0, fmt.Errorf("timeout string is too short: %q", s)
	}
	d, ok := timeoutUnitToDuration(s[size-1])
	if !ok {
		return 0, fmt.Errorf("timeout unit is not recognized: %q", s)
	}
	t, err := strconv.ParseInt(s[:size-1], 10, 64)
	if err != nil {
		return 0, err
	}
	return d * time.Duration(t), nil
}

func timeoutUnitToDuration(u uint8) (d time.Duration, ok bool) {
	switch u {
	case 'H':
		return time.Hour, true
	case 'M':
		return time.Minute, true
	case 'S':
		return time.Second, true
	case 'm':
		return time.Millisecond, true
	case 'u':
		return time.Microsecond, true
	case 'n':
		return time.Nanosecond, true
	default:
	}
	return
}

// isPermanentHTTPHeader checks whether hdr belongs to the list of
// permanent request headers maintained by IANA.
// http://www.iana.org/assignments/message-headers/message-headers.xml
func isPermanentHTTPHeader(hdr string) bool {
	switch hdr {
	case
		"Accept",
		"Accept-Charset",
		"Accept-Language",
		"Accept-Ranges",
		"Authorization",
		"Cache-Control",
		"Content-Type",
		"Cookie",
		"Date",
		"Expect",
		"From",
		"Host",
		"If-Match",
		"If-Modified-Since",
		"If-None-Match",
		"If-Schedule-Tag-Match",
		"If-Unmodified-Since",
		"Max-Forwards",
		"Origin",
		"Pragma",
		"Referer",
		"User-Agent",
		"Via",
		"Warning":
		return true
	}
	return false
}

// RPCMethod returns the method string for the server context. The returned
// string is in the format of "/package.service/method".
func RPCMethod(ctx context.Context) (string, bool) {
	m := ctx.Value(rpcMethodKey{})
	if m == nil {
		return "", false
	}
	ms, ok := m.(string)
	if !ok {
		return "", false
	}
	return ms, true
}

func withRPCMethod(ctx context.Context, rpcMethodName string) context.Context {
	return context.WithValue(ctx, rpcMethodKey{}, rpcMethodName)
}

type (
	rpcMethodKey       struct{}
	httpPathPatternKey struct{}
)

// HTTPPathPattern returns the HTTP path pattern string relating to the HTTP handler, if one exists.
// The format of the returned string is defined by the google.api.http path template type.
func HTTPPathPattern(ctx context.Context) (string, bool) {
	m := ctx.Value(httpPathPatternKey{})
	if m == nil {
		return "", false
	}
	ms, ok := m.(string)
	if !ok {
		return "", false
	}
	return ms, true
}

func withHTTPPathPattern(ctx context.Context, httpPathPattern string) context.Context {
	return context.WithValue(ctx, httpPathPatternKey{}, httpPathPattern)
}

func AnnotateContext(ctx context.Context, req *http.Request, rpcMethodName string, options ...AnnotateContextOption) (context.Context, error) {
	ctx, md, err := annotateContext(ctx, req, rpcMethodName, options...)
	if err != nil {
		return nil, err
	}
	if md == nil {
		return ctx, nil
	}

	return metadata.NewOutgoingContext(ctx, md), nil
}

type AnnotateContextOption func(ctx context.Context) context.Context

func annotateContext(ctx context.Context, req *http.Request, rpcMethodName string, options ...AnnotateContextOption) (context.Context, metadata.MD, error) {
	ctx = withRPCMethod(ctx, rpcMethodName)
	for _, o := range options {
		ctx = o(ctx)
	}
	var pairs []string
	timeout := time.Second * 2
	if tm := req.Header.Get(metadataGrpcTimeout); tm != "" {
		var err error
		timeout, err = timeoutDecode(tm)
		if err != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "invalid grpc-timeout: %s", tm)
		}
	}

	for key, vals := range req.Header {
		key = textproto.CanonicalMIMEHeaderKey(key)
		for _, val := range vals {
			// For backwards-compatibility, pass through 'authorization' header with no prefix.
			if key == "Authorization" {
				pairs = append(pairs, "authorization", val)
			}
		}
	}

	if host := req.Header.Get(xForwardedHost); host != "" {
		pairs = append(pairs, strings.ToLower(xForwardedHost), host)
	} else if req.Host != "" {
		pairs = append(pairs, strings.ToLower(xForwardedHost), req.Host)
	}

	if addr := req.RemoteAddr; addr != "" {
		if remoteIP, _, err := net.SplitHostPort(addr); err == nil {
			if fwd := req.Header.Get(xForwardedFor); fwd == "" {
				pairs = append(pairs, strings.ToLower(xForwardedFor), remoteIP)
			} else {
				pairs = append(pairs, strings.ToLower(xForwardedFor), fmt.Sprintf("%s, %s", fwd, remoteIP))
			}
		}
	}

	if timeout != 0 {
		//nolint:govet  // The context outlives this function
		ctx, _ = context.WithTimeout(ctx, timeout)
	}
	if len(pairs) == 0 {
		return ctx, nil, nil
	}
	return ctx, metadata.Pairs(pairs...), nil
}

// responseBody interface contains method for getting field for marshaling to the response body
// this method is generated for response struct from the value of `response_body` in the `google.api.HttpRule`
type responseBody interface {
	XXX_ResponseBody() interface{}
}

// Delimited defines the streaming delimiter.
type Delimited interface {
	// Delimiter returns the record separator for the stream.
	Delimiter() []byte
}

// HTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	grpclog.Infof("Unknown gRPC error code: %v", code)
	return http.StatusInternalServerError
}

func requestAcceptsTrailers(req *http.Request) bool {
	te := req.Header.Get("TE")
	return strings.Contains(strings.ToLower(te), "trailers")
}

func handleForwardResponseOptions(ctx context.Context, w http.ResponseWriter, resp proto.Message, opts []func(context.Context, http.ResponseWriter, proto.Message) error) error {
	if len(opts) == 0 {
		return nil
	}
	for _, opt := range opts {
		if err := opt(ctx, w, resp); err != nil {
			grpclog.Infof("Error handling ForwardResponseOptions: %v", err)
			return err
		}
	}
	return nil
}

func handleForwardResponseStreamError(ctx context.Context, wroteHeader bool, w http.ResponseWriter, req *http.Request, err error) {
	if !wroteHeader {
		w.WriteHeader(HTTPStatusFromCode(500))
	}
}

func errorChunk(st *status.Status) map[string]proto.Message {
	return map[string]proto.Message{"error": st.Proto()}
}

// MetadataHeaderPrefix is the http prefix that represents custom metadata
// parameters to or from a gRPC call.
const MetadataHeaderPrefix = "Grpc-Metadata-"

// MetadataPrefix is prepended to permanent HTTP header keys (as specified
// by the IANA) when added to the gRPC context.
const MetadataPrefix = "grpcgateway-"

// MetadataTrailerPrefix is prepended to gRPC metadata as it is converted to
// HTTP headers in a response handled by grpc-gateway
const MetadataTrailerPrefix = "Grpc-Trailer-"

const metadataGrpcTimeout = "Grpc-Timeout"
const metadataHeaderBinarySuffix = "-Bin"

const xForwardedFor = "X-Forwarded-For"
const xForwardedHost = "X-Forwarded-Host"

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

type serverMetadataKey struct{}

// ServerMetadata consists of metadata sent from gRPC server.
type ServerMetadata struct {
	HeaderMD  metadata.MD
	TrailerMD metadata.MD
}

// ServerMetadataFromContext returns the ServerMetadata in ctx
func ServerMetadataFromContext(ctx context.Context) (md ServerMetadata, ok bool) {
	md, ok = ctx.Value(serverMetadataKey{}).(ServerMetadata)
	return
}

func ForwardResponseMessage(ctx context.Context, w http.ResponseWriter, req *http.Request, resp proto.Message, opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	md, ok := ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
	}

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.
	doForwardTrailers := requestAcceptsTrailers(req)

	if doForwardTrailers {
		handleForwardResponseTrailerHeader(w, md)
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	handleForwardResponseTrailerHeader(w, md)

	xerror.Panic(handleForwardResponseOptions(ctx, w, resp, opts))

	var buf []byte
	var err error
	if rb, ok := resp.(responseBody); ok {
		buf, err = json.Marshal(rb.XXX_ResponseBody())
	} else {
		buf, err = json.Marshal(resp)
	}
	if err != nil {
		grpclog.Infof("Marshal error: %v", err)
		xerror.Panic(err)
		return
	}

	if _, err = w.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

	if doForwardTrailers {
		handleForwardResponseTrailer(w, md)
	}
}

func ForwardResponseStream(ctx context.Context, w http.ResponseWriter, req *http.Request, recv func() (proto.Message, error), opts ...func(context.Context, http.ResponseWriter, proto.Message) error) {
	f, ok := w.(http.Flusher)
	if !ok {
		grpclog.Infof("Flush not supported in %T", w)
		http.Error(w, "unexpected type of web server", http.StatusInternalServerError)
		return
	}

	md, ok := ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
	_ = md

	w.Header().Set("Transfer-Encoding", "chunked")
	xerror.Panic(handleForwardResponseOptions(ctx, w, nil, opts))

	var delimiter []byte
	delimiter = []byte("\n")

	var wroteHeader bool
	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			handleForwardResponseStreamError(ctx, wroteHeader, w, req, err)
			return
		}
		if err := handleForwardResponseOptions(ctx, w, resp, opts); err != nil {
			handleForwardResponseStreamError(ctx, wroteHeader, w, req, err)
			return
		}

		if !wroteHeader {
			w.Header().Set("Content-Type", "")
		}

		var buf []byte
		httpBody, isHTTPBody := resp.(*httpbody.HttpBody)
		switch {
		case resp == nil:
			buf, err = json.Marshal(errorChunk(status.New(codes.Internal, "empty response")))
		case isHTTPBody:
			buf = httpBody.GetData()
		default:
			result := map[string]interface{}{"result": resp}
			if rb, ok := resp.(responseBody); ok {
				result["result"] = rb.XXX_ResponseBody()
			}

			buf, err = json.Marshal(result)
		}

		if err != nil {
			grpclog.Infof("Failed to marshal response chunk: %v", err)
			handleForwardResponseStreamError(ctx, wroteHeader, w, req, err)
			return
		}
		if _, err = w.Write(buf); err != nil {
			grpclog.Infof("Failed to send response chunk: %v", err)
			return
		}
		wroteHeader = true
		if _, err = w.Write(delimiter); err != nil {
			grpclog.Infof("Failed to send delimiter chunk: %v", err)
			return
		}
		f.Flush()
	}
}

func init() {
	//var filter_UserService_ListUsers_0=&utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
	//if err := req.ParseForm(); err != nil {
	//	return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	//}
	//if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_UserService_ListUsers_0); err != nil {
	//	return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	//}
}
