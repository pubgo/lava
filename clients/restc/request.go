package restc

import (
	"net/http"

	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/types"
)

var _ types.Request = (*Request)(nil)

// Metrics contains the metrics about each request
type Metrics struct {
	// Failures is the number of failed requests
	Failures int
	// Retries is the number of retries for the request
	Retries int
	// DrainErrors is number of errors occured in draining response body
	DrainErrors int
}

type Request struct {
	clientTrace *clientTrace
	req         *http.Request
	service     string
	ct          string
	cdc         encoding.Codec
	data        []byte
	// Metrics contains the metrics for the request.
	Metrics Metrics
}

func (r *Request) TraceInfo() TraceInfo {
	ct := r.clientTrace

	if ct == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		DNSLookup:     ct.dnsDone.Sub(ct.dnsStart),
		TLSHandshake:  ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		ServerTime:    ct.gotFirstResponseByte.Sub(ct.gotConn),
		IsConnReused:  ct.gotConnInfo.Reused,
		IsConnWasIdle: ct.gotConnInfo.WasIdle,
		ConnIdleTime:  ct.gotConnInfo.IdleTime,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if ct.gotConnInfo.Reused {
		ti.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		ti.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}

func (r *Request) Kind() string          { return Name }
func (r *Request) Codec() encoding.Codec { return r.cdc }
func (r *Request) Client() bool          { return true }
func (r *Request) Service() string       { return r.service }
func (r *Request) Method() string        { return r.req.Method }
func (r *Request) Endpoint() string      { return r.req.RequestURI }
func (r *Request) ContentType() string   { return r.ct }
func (r *Request) Header() types.Header  { return types.Header(r.req.Header) }
func (r *Request) Payload() interface{}  { return r.data }
func (r *Request) Read() ([]byte, error) { return r.data, nil }
func (r *Request) Stream() bool          { return false }
