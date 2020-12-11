package grpclient_pool

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"google.golang.org/grpc"
)

func WithDefaultDialOptions(opts ...grpc.DialOption) []grpc.DialOption {
	return append(opts, withContextDialer())
}

type Dialer struct {
	dialed       bool
	reconnectCnt int
	createdAt    time.Time
}

func withContextDialer() grpc.DialOption {
	d := &Dialer{}

	dialer := newProxyDialer(
		func(ctx context.Context, addr string) (net.Conn, error) {
			if !d.dialed {
				d.dialed = true
				d.createdAt = time.Now()
			}

			network, addr := parseDialTarget(addr)

			log.Printf(
				"[grpc] Dialer - ID: %x, network: %s, addr: %s, reconnect count: %d, created at: %s\n",
				uintptr(unsafe.Pointer(d)), network, addr, d.reconnectCnt, d.createdAt.Format(time.RFC3339Nano))

			d.reconnectCnt++

			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	)

	return grpc.WithContextDialer(dialer)
}

// ----------------------------------------------------------------------------------------------------------------------
var (
	// errDisabled indicates that proxy is disabled for the address.
	errDisabled = errors.New("proxy is disabled for the address")
	// The following variable will be overwritten in the tests.
	httpProxyFromEnvironment = http.ProxyFromEnvironment
)

const (
	proxyAuthHeaderKey = "Proxy-Authorization"
	grpcUA             = "grpc-go/" + grpc.Version
)

// newProxyDialer returns a dialer that connects to proxy first if necessary.
// The returned dialer checks if a proxy is necessary, dial to the proxy with the
// provided dialer, does HTTP CONNECT handshake and returns the connection.
func newProxyDialer(dialer func(context.Context, string) (net.Conn, error)) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (conn net.Conn, err error) {
		var newAddr string
		proxyURL, err := mapAddress(ctx, addr)
		if err != nil {
			if err != errDisabled {
				return nil, err
			}
			newAddr = addr
		} else {
			newAddr = proxyURL.Host
		}

		conn, err = dialer(ctx, newAddr)
		if err != nil {
			return
		}
		if proxyURL != nil {
			// proxy is disabled if proxyURL is nil.
			conn, err = doHTTPConnectHandshake(ctx, conn, addr, proxyURL)
		}
		return
	}
}

// parseDialTarget returns the network and address to pass to dialer
func parseDialTarget(target string) (net string, addr string) {
	net = "tcp"

	m1 := strings.Index(target, ":")
	m2 := strings.Index(target, ":/")

	// handle unix:addr which will fail with url.Parse
	if m1 >= 0 && m2 < 0 {
		if n := target[0:m1]; n == "unix" {
			net = n
			addr = target[m1+1:]
			return net, addr
		}
	}
	if m2 >= 0 {
		t, err := url.Parse(target)
		if err != nil {
			return net, target
		}
		scheme := t.Scheme
		addr = t.Path
		if scheme == "unix" {
			net = scheme
			if addr == "" {
				addr = t.Host
			}
			return net, addr
		}
	}

	return net, target
}

func mapAddress(ctx context.Context, address string) (*url.URL, error) {
	req := &http.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   address,
		},
	}
	url, err := httpProxyFromEnvironment(req)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, errDisabled
	}
	return url, nil
}

func doHTTPConnectHandshake(ctx context.Context, conn net.Conn, backendAddr string, proxyURL *url.URL) (_ net.Conn, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: backendAddr},
		Header: map[string][]string{"User-Agent": {grpcUA}},
	}
	if t := proxyURL.User; t != nil {
		u := t.Username()
		p, _ := t.Password()
		req.Header.Add(proxyAuthHeaderKey, "Basic "+basicAuth(u, p))
	}

	if err := sendHTTPRequest(ctx, req, conn); err != nil {
		return nil, fmt.Errorf("failed to write the HTTP request: %v", err)
	}

	r := bufio.NewReader(conn)
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, fmt.Errorf("reading server HTTP response: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("failed to do connect handshake, status code: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to do connect handshake, response: %q", dump)
	}

	return &bufConn{Conn: conn, r: r}, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func sendHTTPRequest(ctx context.Context, req *http.Request, conn net.Conn) error {
	req = req.WithContext(ctx)
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("failed to write the HTTP request: %v", err)
	}
	return nil
}

// To read a response from a net.Conn, http.ReadResponse() takes a bufio.Reader.
// It's possible that this reader reads more than what's need for the response and stores
// those bytes in the buffer.
// bufConn wraps the original net.Conn and the bufio.Reader to make sure we don't lose the
// bytes in the buffer.
type bufConn struct {
	net.Conn
	r io.Reader
}

func (c *bufConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}
